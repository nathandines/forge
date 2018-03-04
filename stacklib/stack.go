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
	StackInfo       *cloudformation.Stack
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

// GetStackInfo populates the StackInfo for this object from the existing stack
// found in the environment
func (s *Stack) GetStackInfo() (err error) {
	var stackName *string
	if s.StackInfo != nil {
		stackName = s.StackInfo.StackId
	} else {
		stackName = &s.StackName
	}
	if stackOut, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: stackName,
	}); err == nil {
		s.StackInfo = stackOut.Stacks[0]
	} else {
		return err
	}
	return
}
