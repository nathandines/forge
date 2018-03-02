package stacklib

import "github.com/aws/aws-sdk-go/service/cloudformation"

// Destroy will delete the stack
func (s *Stack) Destroy() (err error) {
	_, err = cfn.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: &s.StackName,
	})
	return
}
