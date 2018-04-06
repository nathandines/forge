package stacklib

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var readFile = ioutil.ReadFile

// DeployOut provides a controlled format for information to be passed out of
// the Deploy function
type DeployOut struct {
	Message string
}

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() (output DeployOut, err error) {
	if err := verifyStackName(s.StackName); err != nil {
		return output, err
	}
	templateBody, err := readFile(s.TemplateFile)
	if err != nil {
		return output, err
	}
	cfn.ValidateTemplate(
		&cloudformation.ValidateTemplateInput{
			TemplateBody: aws.String(string(templateBody)),
		},
	)

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
	if s.StackInfo == nil {
		_, err := cfn.CreateStack(
			&cloudformation.CreateStackInput{
				StackName:    aws.String(s.StackName),
				TemplateBody: aws.String(string(templateBody)),
				OnFailure:    aws.String("DELETE"),
			},
		)
		if err != nil {
			return output, err
		}
	} else {
		_, err := cfn.UpdateStack(
			&cloudformation.UpdateStackInput{
				StackName:    aws.String(s.StackID),
				TemplateBody: aws.String(string(templateBody)),
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
