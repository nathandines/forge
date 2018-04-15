package forgelib

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

// Stack represents the attributes of a stack deployment, including the AWS
// parameters, and local resources which represent what needs to be deployed
type Stack struct {
	ParametersBody  string
	ProjectManifest string
	CfnRoleName     string
	StackID         string
	StackInfo       *cloudformation.Stack
	StackName       string
	StackPolicyFile string
	TagsBody        string
	TemplateBody    string
}

var cfnClient cloudformationiface.CloudFormationAPI // CloudFormation service
var stsClient stsiface.STSAPI                       // STS Service

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	cfnClient = cloudformation.New(sess)
	stsClient = sts.New(sess)
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
