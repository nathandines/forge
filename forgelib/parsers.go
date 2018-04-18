package forgelib

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

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
		envVarSub, err := parseEnvironmentVariables(parsedVal)
		if err != nil {
			return output, err
		}
		output = append(output, &cloudformation.Tag{
			Key:   aws.String(k),
			Value: aws.String(envVarSub),
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
		return output, fmt.Errorf("Parameters must be a basic key-value object")
	}
	for k, v := range parsedMap {
		var parsedVal string
		if err := valueToString(v, &parsedVal, true, true); err != nil {
			return output, fmt.Errorf("Invalid Parameter %s: %s", k, err)
		}
		envVarSub, err := parseEnvironmentVariables(parsedVal)
		if err != nil {
			return output, err
		}
		output = append(output, &cloudformation.Parameter{
			ParameterKey:   aws.String(k),
			ParameterValue: aws.String(envVarSub),
		})
	}
	return output, err
}

func parseEnvironmentVariables(input string) (string, error) {
	funcMap := template.FuncMap{
		"env": func(input string) (string, error) {
			if v, present := os.LookupEnv(input); present {
				return v, nil
			}
			return "", fmt.Errorf("Environment variable by the name \"%s\" is not defined", input)
		},
	}

	envTemplate, err := template.New("envTemplate").Funcs(funcMap).Parse(input)
	if err != nil {
		return "", err
	}

	var outputBuffer bytes.Buffer
	if err := envTemplate.Execute(&outputBuffer, nil); err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}
