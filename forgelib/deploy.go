package forgelib

import (
	"encoding/json"
	"fmt"

	//	"github.com/ghodss/yaml"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/pkg/errors"
)

// DeployOut provides a controlled format for information to be passed out of
// the Deploy function
type DeployOut struct {
	Message string
}

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() (output DeployOut, err error) {

	var validateConfig cloudformation.ValidateTemplateInput

	if s.TemplateUrl != "" {
		validateConfig = cloudformation.ValidateTemplateInput{TemplateURL: aws.String(s.TemplateUrl)}
	}

	if s.TemplateBody != "" {
		validateConfig = cloudformation.ValidateTemplateInput{TemplateBody: aws.String(s.TemplateBody)}
	}

	validationResult, err := cfnClient.ValidateTemplate(&validateConfig)
	if err != nil {
		return output, errors.Wrap(err, "Failed to Validate Template")
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
			return output, err
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
			return output, errors.Wrap(err, "Failed to get Template Summary")
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
				return output, err
			}
		} else {
			return output, errors.Wrap(err, "Error in GetStackInfo")
		}
	}

	var tags []*cloudformation.Tag
	if s.TagsBody != "" {
		tags, err = parseTags(s.TagsBody)
		if err != nil {
			return output, err
		}
	} else if s.StackInfo != nil {
		tags = s.StackInfo.Tags
	}

	var roleARN *string
	if s.CfnRoleName != "" {
		roleARNString, err := roleARNFromName(s.CfnRoleName)
		if err != nil {
			return output, err
		}
		roleARN = &roleARNString
	}

	//var inputStackPolicy *string
	//if s.StackPolicyBody != "" {
	//	jsonStackPolicy, err := yaml.YAMLToJSON([]byte(s.StackPolicyBody))
	//	if err != nil {
	//		return output, errors.Wrap(err, "Unable to convert template policy from YAML to JSON")
	//	}
	//	jsonStackPolicyString := string(jsonStackPolicy)
	//	inputStackPolicy = &jsonStackPolicyString
	//}

	createConfig := cloudformation.CreateStackInput{
		StackName:    aws.String(s.StackName),
		OnFailure:    aws.String("DELETE"),
		Capabilities: validationResult.Capabilities,
		Tags:         tags,
		Parameters:   inputParams,
		RoleARN:      roleARN,
		//StackPolicyBody:             inputStackPolicy,
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
		if err != nil {
			return output, errors.Wrap(err, "Failed to Create Stack")
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
				return output, errors.Wrap(err, "Failed to Update Termination Protection")
			}
		}

		// If stack pnput --stack-policy-recurse and inputStackPolicy
		//	Create a function that takes a stack policy and stack template

		updateConfig := cloudformation.UpdateStackInput{
			StackName:    aws.String(s.StackID),
			Capabilities: validationResult.Capabilities,
			Tags:         tags,
			Parameters:   inputParams,
			RoleARN:      roleARN,
			//StackPolicyBody: inputStackPolicy,
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
					return DeployOut{Message: noUpdatesErr}, nil
				}
			}
			return output, errors.Wrap(err, "Failed to Update Stack")
		}
	}
	return
}
