package stacklib

import "fmt"

// Destroy will delete the stack
func (s *Stack) Destroy() {
	fmt.Println(s.StackName, s.ProjectManifest)
}
