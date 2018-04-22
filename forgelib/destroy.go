package forgelib

import (
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// Destroy will delete the stack
func (s *Stack) Destroy() (err error) {
	if s.StackID == "" {
		return errorNoStackID
	}

	var roleARN *string
	if s.CfnRoleName != "" {
		roleARNString, err := roleARNFromName(s.CfnRoleName)
		if err != nil {
			return err
		}
		roleARN = &roleARNString
	}

	// Delete stack by Stack ID. This removes the risk of deleting a stack with
	// the same name which was created since this was previously executed. The
	// `Stack` object should always refer to the exact same stack, be it created
	// or deleted
	_, err = cfnClient.DeleteStack(
		&cloudformation.DeleteStackInput{
			StackName: &s.StackID,
			RoleARN:   roleARN,
		},
	)
	return
}
