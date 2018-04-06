package stacklib

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var readFile = ioutil.ReadFile

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() (err error) {
	if err := verifyStackName(s.StackName); err != nil {
		return err
	}
	templateBody, err := readFile(s.TemplateFile)
	if err != nil {
		return err
	}
	cfn.ValidateTemplate(
		&cloudformation.ValidateTemplateInput{
			TemplateBody: aws.String(string(templateBody)),
		},
	)

	if s.GetStackInfo(); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Message() {
			case fmt.Sprintf("Stack with id %s does not exist", s.StackID),
				fmt.Sprintf("Stack with id %s does not exist", s.StackName):
			default:
				return err
			}
		} else {
			return err
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
			return err
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
				if awsErr.Message() != noUpdatesErr {
					return err
				}
				fmt.Println(noUpdatesErr)
			} else {
				return err
			}
		}
	}
	return
}
