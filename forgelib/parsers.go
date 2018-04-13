package forgelib

import (
	"bytes"
	"fmt"
	"reflect"
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
		if err := valueToString(v, &parsedVal, false, true); err != nil {
			return output, fmt.Errorf("Invalid Tag %s: %s", k, err)
		}
		output = append(output, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(parsedVal),
		})
	}
	return output, err
}

func parseParameters(input string) (output []*cloudformation.Parameter, err error) {
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
		if err := valueToString(v, &parsedVal, true, true); err != nil {
			return output, fmt.Errorf("Invalid Parameter %s: %s", k, err)
		}
		output = append(output, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(parsedVal),
		})
	}
	return output, err
}

func valueToString(v interface{}, dest *string, allowSlices bool, allowCommas bool) error {
	if allowSlices && reflect.TypeOf(v).Kind() == reflect.Slice {
		var buffer bytes.Buffer
		vv := reflect.ValueOf(v)
		for i := 0; i < vv.Len(); i++ {
			var thisOutput string
			if err := valueToString(vv.Index(i).Interface(), &thisOutput, false, false); err != nil {
				return err
			}
			if buffer.Len() > 0 {
				buffer.WriteByte(',')
			}
			buffer.WriteString(thisOutput)
		}
		*dest = buffer.String()
		return nil
	}

	switch vv := v.(type) {
	case string:
		for _, b := range vv {
			if !allowCommas && b == ',' {
				return fmt.Errorf("Commas not allowed in list values")
			}
		}
		*dest = vv
	// float64 also captures integers from YAML/JSON
	case float64:
		*dest = strconv.FormatFloat(vv, 'f', -1, 64)
	case bool:
		*dest = strconv.FormatBool(vv)
	default:
		return fmt.Errorf("Field of type %s is not allowed", reflect.TypeOf(vv).Kind().String())
	}
	return nil
}
