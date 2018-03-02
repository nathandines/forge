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
