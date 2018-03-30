package stacklib

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// Stack represents the attributes of a stack deployment, including the AWS
// paramters, and local resources which represent what needs to be deployed
type Stack struct {
	ParametersFile  string
	ProjectManifest string
	RoleName        string
	StackID         string
	StackInfo       *cloudformation.Stack
	StackName       string
	StackPolicyFile string
	TemplateFile    string
}

var cfn cloudformationiface.CloudFormationAPI // CloudFormation service

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
	if s.StackID != "" {
		stackName = &s.StackID
	} else if s.StackName != "" {
		stackName = &s.StackName
	} else {
		return errorNoStackNameOrID
	}
	if stackOut, err := cfn.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: stackName,
	}); err == nil {
		s.StackInfo = stackOut.Stacks[0]
		if s.StackName == "" {
			s.StackName = *s.StackInfo.StackName
		}
		if s.StackID == "" {
			s.StackID = *s.StackInfo.StackId
		}
	} else {
		return err
	}
	return
}
