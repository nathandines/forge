package stacklib

import "fmt"

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() {
	fmt.Println(s.StackName, s.TemplateFile)
}
