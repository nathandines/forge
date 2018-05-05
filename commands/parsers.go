package commands

import (
	"fmt"
	"strings"
)

func parseParameterOverrideArgs(input []string) (map[string]string, error) {
	output := map[string]string{}
	for _, i := range input {
		inputSlice := strings.Split(i, "=")
		if len(inputSlice) < 2 {
			return output, fmt.Errorf("Parameter override \"%s\" is invalid. Must be of the format \"<key>=<value>\"", i)
		}

		key := inputSlice[0]
		value := strings.Join(inputSlice[1:], "=")

		output[key] = value
	}
	return output, nil
}
