package forgelib

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/minio/minio/pkg/wildcard"
	"github.com/pkg/errors"
)

// DeployOut provides a controlled format for information to be passed out of
// the Deploy function
type DeployOut struct {
	Message string
}

type PolicyDocument struct {
	Statement []StatementEntry
}

type StatementEntry struct {
	Effect      string                         `json:",omitempty"`
	Action      interface{}                    `json:",omitempty"` // May be string or []string
	NotAction   string                         `json:",omitempty"`
	Principal   string                         `json:",omitempty"`
	Resource    string                         `json:",omitempty"`
	NotResource string                         `json:",omitempty"`
	Condition   map[string]map[string][]string `json:",omitempty"`
}

func recursiveSetStackPolicy(stackPolicyBody *string, stackName *string) error {
	stackResourcesInput := cloudformation.DescribeStackResourcesInput{
		StackName: stackName,
	}
	stackResources, err := cfnClient.DescribeStackResources(&stackResourcesInput)
	if err != nil {
		return err
	}

	var stackResouceLogicalNames []*string
	var stackResourcNestedStacks []*string

	for _, stackResource := range stackResources.StackResources {
		if *stackResource.ResourceType == "AWS::CloudFormation::Stack" {
			stackResourcNestedStacks = append(stackResourcNestedStacks, stackResource.PhysicalResourceId)
		}
		stackResouceLogicalNames = append(stackResouceLogicalNames, stackResource.LogicalResourceId)
	}

	var stackPolicy PolicyDocument
	json.Unmarshal([]byte(*stackPolicyBody), &stackPolicy)

	// filter stack policy. remove logical ids not included in this stack
	var filteredStackPolicy PolicyDocument
	for _, statementEntry := range stackPolicy.Statement {
		var logicalNameSplit []string
		if statementEntry.Resource != "" {
			logicalNameSplit = strings.Split(statementEntry.Resource, "/")
		} else if statementEntry.NotResource != "" {
			logicalNameSplit = strings.Split(statementEntry.NotResource, "/")
		}
		logicalNamePattern := logicalNameSplit[len(logicalNameSplit)-1]
		for _, stackResouceLogicalName := range stackResouceLogicalNames {
			if wildcard.MatchSimple(logicalNamePattern, *stackResouceLogicalName) {
				filteredStackPolicy.Statement = append(filteredStackPolicy.Statement, statementEntry)
				break
			}
		}
	}

	jsonStackPolicy, err := json.Marshal(filteredStackPolicy)
	if err != nil {
		return err
	}
	jsonStackPolicyString := string(jsonStackPolicy)

	setStackPolicyInput := cloudformation.SetStackPolicyInput{
		StackName:       stackName,
		StackPolicyBody: &jsonStackPolicyString,
	}

	cfnClient.SetStackPolicy(&setStackPolicyInput)

	// For each nested stack recursiveSetStackPolicy()
	for _, stackResourcNestedStack := range stackResourcNestedStacks {
		err = recursiveSetStackPolicy(stackPolicyBody, stackResourcNestedStack)
		if err != nil {
			return err
		}
	}
	return nil

}

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() (output DeployOut, postActions []func(), err error) {

	var validateConfig cloudformation.ValidateTemplateInput

	if s.TemplateUrl != "" {
		validateConfig = cloudformation.ValidateTemplateInput{TemplateURL: aws.String(s.TemplateUrl)}
	}

	if s.TemplateBody != "" {
		validateConfig = cloudformation.ValidateTemplateInput{TemplateBody: aws.String(s.TemplateBody)}
	}

	validationResult, err := cfnClient.ValidateTemplate(&validateConfig)
	if err != nil {
		return output, postActions, errors.Wrap(err, "Failed to Validate Template")
	}

	// CAPABILITY_AUTO_EXPAND is not discovered during validation so add it here.  These lines may
	// be removed if validate-template is able to discover this capability requirement in the future.
	capabilityCapabilityAutoExpand := cloudformation.CapabilityCapabilityAutoExpand
	validationResult.Capabilities = append(validationResult.Capabilities, &capabilityCapabilityAutoExpand)

	var inputParams []*cloudformation.Parameter
	parsedParameters := []*cloudformation.Parameter{}
	if len(s.ParameterBodies) != 0 {
		parsedParameters, err = parseParameters(s.ParameterBodies)
		if err != nil {
			return output, postActions, err
		}
	}

TEMPLATE_PARAMETERS:
	for i := 0; i < len((*validationResult).Parameters); i++ {
		parameterKey := *(*validationResult).Parameters[i].ParameterKey
		if v, ok := s.ParameterOverrides[parameterKey]; ok {
			param := cloudformation.Parameter{
				ParameterKey:   aws.String(parameterKey),
				ParameterValue: aws.String(v),
			}
			inputParams = append(inputParams, &param)
			continue TEMPLATE_PARAMETERS
		}
		for j := 0; j < len(parsedParameters); j++ {
			if *parsedParameters[j].ParameterKey == parameterKey {
				inputParams = append(inputParams, parsedParameters[j])
				continue TEMPLATE_PARAMETERS
			}
		}
	}

	if s.StackName == "" {
		/*
			If StackName is not defined use Metadata:StackName:<Parameter>:<ParameterValue>

			------ Example with Environment Parameter ----
			  Metadata:
			    StackName:
			      Environment:
			        staging: search-ui-v1-staging
			        production: search-ui-v1-production
		*/
		var templateSummaryConfig cloudformation.GetTemplateSummaryInput

		if s.TemplateUrl != "" {
			templateSummaryConfig = cloudformation.GetTemplateSummaryInput{TemplateURL: aws.String(s.TemplateUrl)}
		}

		if s.TemplateBody != "" {
			templateSummaryConfig = cloudformation.GetTemplateSummaryInput{TemplateBody: aws.String(s.TemplateBody)}
		}

		TemplateSummary, err := cfnClient.GetTemplateSummary(&templateSummaryConfig)
		if err != nil {
			return output, postActions, errors.Wrap(err, "Failed to get Template Summary")
		}

		type Metadata struct {
			StackName map[string]map[string]string
		}
		myMetadata := Metadata{}

		if TemplateSummary.Metadata != nil {
			json.Unmarshal([]byte(*TemplateSummary.Metadata), &myMetadata)

			for k, v := range myMetadata.StackName {
				for _, pv := range inputParams {
					if *pv.ParameterKey == k {
						s.StackName = v[*pv.ParameterValue]
					}
				}
			}
		}
	}

	if err := s.GetStackInfo(); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Message() {
			case fmt.Sprintf("Stack with id %s does not exist", s.StackID),
				fmt.Sprintf("Stack with id %s does not exist", s.StackName):
			default:
				return output, postActions, err
			}
		} else {
			return output, postActions, errors.Wrap(err, "Error in GetStackInfo")
		}
	}

	var tags []*cloudformation.Tag
	if s.TagsBody != "" {
		tags, err = parseTags(s.TagsBody)
		if err != nil {
			return output, postActions, err
		}
	} else if s.StackInfo != nil {
		tags = s.StackInfo.Tags
	}

	var roleARN *string
	if s.CfnRoleName != "" {
		roleARNString, err := roleARNFromName(s.CfnRoleName)
		if err != nil {
			return output, postActions, err
		}
		roleARN = &roleARNString
	}

	createConfig := cloudformation.CreateStackInput{
		StackName:                   aws.String(s.StackName),
		OnFailure:                   aws.String("DELETE"),
		Capabilities:                validationResult.Capabilities,
		Tags:                        tags,
		Parameters:                  inputParams,
		RoleARN:                     roleARN,
		EnableTerminationProtection: aws.Bool(s.TerminationProtection),
	}

	if s.TemplateUrl != "" {
		createConfig.TemplateURL = aws.String(s.TemplateUrl)
	}

	if s.TemplateBody != "" {
		createConfig.TemplateBody = aws.String(s.TemplateBody)
	}

	if s.StackInfo == nil {
		createOut, err := cfnClient.CreateStack(&createConfig)

		if s.StackPolicyBody != "" {
			postActions = append(postActions, func(stackPolicyBody string, stackId string) func() {
				return func() {
					recursiveSetStackPolicy(&stackPolicyBody, &stackId)
				}
			}(s.StackPolicyBody, *createOut.StackId))
		}

		if err != nil {
			return output, postActions, errors.Wrap(err, "Failed to Create Stack")
		}
		s.StackID = *createOut.StackId

	} else {
		// Only SET termination protection, do not remove
		if s.StackInfo.EnableTerminationProtection != nil &&
			!*s.StackInfo.EnableTerminationProtection &&
			s.TerminationProtection {
			_, err := cfnClient.UpdateTerminationProtection(
				&cloudformation.UpdateTerminationProtectionInput{
					EnableTerminationProtection: aws.Bool(s.TerminationProtection),
					StackName:                   aws.String(s.StackID),
				},
			)
			if err != nil {
				return output, postActions, errors.Wrap(err, "Failed to Update Termination Protection")
			}
		}

		if s.StackPolicyBody != "" {
			err = recursiveSetStackPolicy(&s.StackPolicyBody, &s.StackID)
			if err != nil {
				return output, postActions, errors.Wrap(err, "Failed to Set Stack Policy")
			}
		}

		updateConfig := cloudformation.UpdateStackInput{
			StackName:    aws.String(s.StackID),
			Capabilities: validationResult.Capabilities,
			Tags:         tags,
			Parameters:   inputParams,
			RoleARN:      roleARN,
		}

		if s.TemplateUrl != "" {
			updateConfig.TemplateURL = aws.String(s.TemplateUrl)
		}

		if s.TemplateBody != "" {
			updateConfig.TemplateBody = aws.String(s.TemplateBody)
		}
		_, err := cfnClient.UpdateStack(&updateConfig)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				noUpdatesErr := "No updates are to be performed."
				if awsErr.Message() == noUpdatesErr {
					return DeployOut{Message: noUpdatesErr}, postActions, nil
				}
			}
			return output, postActions, errors.Wrap(err, "Failed to Update Stack")
		}
	}
	return
}
