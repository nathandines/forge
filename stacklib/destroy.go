package stacklib

import "github.com/aws/aws-sdk-go/service/cloudformation"

// Destroy will delete the stack
func (s *Stack) Destroy() (err error) {
	if err := s.GetStackInfo(); err != nil {
		return err
	}
	_, err = cfn.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: s.StackInfo.StackId,
	})
	return
}
