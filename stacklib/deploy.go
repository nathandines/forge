package stacklib

// Deploy will create or update the stack (depending on its current state)
func (s *Stack) Deploy() (err error) {
	if err := verifyStackName(s.StackName); err != nil {
		return err
	}
	return
}
