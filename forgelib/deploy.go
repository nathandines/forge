package forgelib

import (
	"fmt"

	"github.com/ghodss/yaml"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// DeployOut provides a controlled format for information to be passed out of
// the Deploy function
type DeployOut struct {
	Message string
}

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() (output DeployOut, err error) {
	validationResult, err := cfnClient.ValidateTemplate(
		&cloudformation.ValidateTemplateInput{
			TemplateBody: aws.String(s.TemplateBody),
		},
	)
	if err != nil {
		return output, err
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
			return output, err
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

	var inputParams []*cloudformation.Parameter
	if s.ParametersBody != "" {
		parameters, err := parseParameters(s.ParametersBody)
		if err != nil {
			return output, err
		}

	INPUT_PARAMETERS:
		for i := 0; i < len(parameters); i++ {
			for j := 0; j < len((*validationResult).Parameters); j++ {
				if *parameters[i].ParameterKey == *(*validationResult).Parameters[j].ParameterKey {
					inputParams = append(inputParams, parameters[i])
					continue INPUT_PARAMETERS
				}
			}
		}
	}

	var roleARN string
	if s.CfnRoleName != "" {
		roleARN, err = roleARNFromName(s.CfnRoleName)
		if err != nil {
			return output, err
		}
	}

	var inputStackPolicy string
	if s.StackPolicyBody != "" {
		jsonStackPolicy, err := yaml.YAMLToJSON([]byte(s.StackPolicyBody))
		if err != nil {
			return output, err
		}
		inputStackPolicy = string(jsonStackPolicy)
	}

	if s.StackInfo == nil {
		createOut, err := cfnClient.CreateStack(
			&cloudformation.CreateStackInput{
				StackName:       aws.String(s.StackName),
				TemplateBody:    aws.String(s.TemplateBody),
				OnFailure:       aws.String("DELETE"),
				Capabilities:    validationResult.Capabilities,
				Tags:            tags,
				Parameters:      inputParams,
				RoleARN:         &roleARN,
				StackPolicyBody: &inputStackPolicy,
			},
		)
		if err != nil {
			return output, err
		}
		s.StackID = *createOut.StackId
	} else {
		_, err := cfnClient.UpdateStack(
			&cloudformation.UpdateStackInput{
				StackName:       aws.String(s.StackID),
				TemplateBody:    aws.String(s.TemplateBody),
				Capabilities:    validationResult.Capabilities,
				Tags:            tags,
				Parameters:      inputParams,
				RoleARN:         &roleARN,
				StackPolicyBody: &inputStackPolicy,
			},
		)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				noUpdatesErr := "No updates are to be performed."
				if awsErr.Message() == noUpdatesErr {
					return DeployOut{Message: noUpdatesErr}, nil
				}
			}
			return output, err
		}
	}
	return
}
