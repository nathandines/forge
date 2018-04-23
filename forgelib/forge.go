package forgelib

import "github.com/aws/aws-sdk-go/service/cloudformation"

// Stack represents the attributes of a stack deployment, including the AWS
// parameters, and local resources which represent what needs to be deployed
type Stack struct {
	ParametersBody        string
	ProjectManifest       string
	CfnRoleName           string
	StackID               string
	StackInfo             *cloudformation.Stack
	StackName             string
	StackPolicyBody       string
	TagsBody              string
	TemplateBody          string
	TerminationProtection bool
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
	stackOut, err := cfnClient.DescribeStacks(&cloudformation.DescribeStacksInput{StackName: stackName})
	if err != nil {
		return err
	}
	s.StackInfo = stackOut.Stacks[0]
	if s.StackName == "" {
		s.StackName = *s.StackInfo.StackName
	}
	if s.StackID == "" {
		s.StackID = *s.StackInfo.StackId
	}
	return
}
