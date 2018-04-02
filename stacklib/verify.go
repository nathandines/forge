package stacklib

import (
	"fmt"
	"regexp"
)

func verifyStackName(stackName string) (err error) {
	regexPattern := "^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$"
	if len(stackName) < 1 || len(stackName) > 128 {
		return fmt.Errorf(
			"stackName must be between 1 and 128 characters",
		)
	}

	result, err := regexp.MatchString(regexPattern, stackName)
	if err != nil {
		return err
	} else if !result {
		return fmt.Errorf(
			"stackName can only contain alphanumeric characters and hyphens, and cannot start or "+
				"end with a hypen (Regex: '%s')",
			regexPattern,
		)
	}
	return
}
