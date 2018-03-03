package stacklib

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Stack represents the attributes of a stack deployment, including the AWS
// paramters, and local resources which represent what needs to be deployed
type Stack struct {
	ParametersFile  string
	ProjectManifest string
	RoleName        string
	StackID         string
	StackName       string
	StackPolicyFile string
	TemplateFile    string
}

var cfn *cloudformation.CloudFormation // CloudFormation service

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	cfn = cloudformation.New(sess)
}

// GetStackID populates the StackID for this object from the existing stack
// found in the environment
func (s *Stack) GetStackID() (err error) {
	if stackOut, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: &s.StackName,
	}); err == nil {
		s.StackID = *stackOut.Stacks[0].StackId
	} else {
		return err
	}
	return
}
