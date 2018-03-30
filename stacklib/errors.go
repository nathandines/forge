package stacklib

import "fmt"

var errorNoStackID = fmt.Errorf("StackID must be defined. Hint: Use GetStackInfo() helper function")
var errorNoStackNameOrID = fmt.Errorf("StackName or StackID must be defined")
