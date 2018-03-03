package stacklib

import "github.com/aws/aws-sdk-go/service/cloudformation"

// Destroy will delete the stack
func (s *Stack) Destroy() (err error) {
	if s.StackID == "" {
		if err := s.GetStackID(); err != nil {
			return err
		}
	}
	_, err = cfn.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: &s.StackName,
	})
	return
}
