package stacklib

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/ghodss/yaml"
)

func parseTags(input string) (output []*cloudformation.Tag, err error) {
	var parsedInput interface{}
	if err := yaml.Unmarshal([]byte(input), &parsedInput); err != nil {
		return output, err
	}
	parsedMap, ok := parsedInput.(map[string]interface{})
	if !ok {
		return output, fmt.Errorf("Tags must be a basic key-value object")
	}
	for k, v := range parsedMap {
		var parsedVal string
		switch vv := v.(type) {
		case string:
			parsedVal = vv
		// float64 also captures integers from YAML/JSON
		case float64:
			parsedVal = strconv.FormatFloat(vv, 'f', -1, 64)
		case bool:
			parsedVal = strconv.FormatBool(vv)
		default:
			return output, fmt.Errorf("Tag values must only be string, integer, or boolean")
		}
		output = append(output, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(parsedVal),
		})
	}
	return output, err
}
