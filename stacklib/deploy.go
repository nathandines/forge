package stacklib

import (
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
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
	cfn.CreateStack(
		&cloudformation.CreateStackInput{
			StackName:    aws.String(s.StackName),
			TemplateBody: aws.String(string(templateBody)),
		},
	)
	return
}
