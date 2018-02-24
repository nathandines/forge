package stacklib

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
