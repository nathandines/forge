package forgelib

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestParseTags(t *testing.T) {
	cases := []struct {
		input        string
		expectedTags []*cloudformation.Tag
		envVars      map[string]string
	}{
		// JSON String
		{
			input: `{"ThisKey":"ThisValue"}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("ThisValue"),
				},
			},
		},
		// YAML String
		{
			input: "---\nThisKey: ThisValue",
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("ThisValue"),
				},
			},
		},
		// JSON Integer
		{
			input: `{"ThisKey":123}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("123"),
				},
			},
		},
		// YAML Integer
		{
			input: "---\nThisKey: 123",
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("123"),
				},
			},
		},
		// JSON Float
		{
			input: `{"ThisKey": 3.1415926535897932384626433832795028841971}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("3.141592653589793"),
				},
			},
		},
		// YAML Float
		{
			input: "---\nThisKey: 3.1415926535897932384626433832795028841971",
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("3.141592653589793"),
				},
			},
		},
		// JSON Bool
		{
			input: `{"ThisKey":true}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("true"),
				},
			},
		},
		// YAML Bool
		{
			input: "---\nThisKey: true",
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("true"),
				},
			},
		},
		// JSON Multi-value
		{
			input: `{
				"String": "Foobar",
                "Int": 123,
                "Float": 123.456,
				"Boolean": true
			}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("String"),
					Value: aws.String("Foobar"),
				},
				{
					Key:   aws.String("Int"),
					Value: aws.String("123"),
				},
				{
					Key:   aws.String("Float"),
					Value: aws.String("123.456"),
				},
				{
					Key:   aws.String("Boolean"),
					Value: aws.String("true"),
				},
			},
		},
		// YAML Multi-value
		{
			input: `---
                String: Foobar
                Int: 123
                Float: 123.456
                Boolean: true`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("String"),
					Value: aws.String("Foobar"),
				},
				{
					Key:   aws.String("Int"),
					Value: aws.String("123"),
				},
				{
					Key:   aws.String("Float"),
					Value: aws.String("123.456"),
				},
				{
					Key:   aws.String("Boolean"),
					Value: aws.String("true"),
				},
			},
		},
		// JSON String, with env variables
		{
			input: `{"ThisKey":"ThisValue-{{ env \"TEST_VAR1\"}}"}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("ThisValue-VALUE_HERE"),
				},
			},
			envVars: map[string]string{"TEST_VAR1": "VALUE_HERE"},
		},
		// YAML String, with env variables
		{
			input: "---\nThisKey: ThisValue-{{ env \"TEST_VAR1\"}}",
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("ThisValue-VALUE_HERE"),
				},
			},
			envVars: map[string]string{"TEST_VAR1": "VALUE_HERE"},
		},
	}

	for i, c := range cases {
		oldValueMap := map[string]string{}
		for k, v := range c.envVars {
			if oldValue, present := os.LookupEnv(k); present {
				oldValueMap[k] = oldValue
				defer os.Setenv(k, oldValueMap[k])
			}
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}

		var tags []*cloudformation.Tag
		tags, err := parseTags(c.input)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

	EXPECTED_TAG:
		for j := 0; j < len(c.expectedTags); j++ {
			e := struct{ Key, Value string }{
				*c.expectedTags[j].Key,
				*c.expectedTags[j].Value,
			}
			for k := 0; k < len(tags); k++ {
				g := struct{ Key, Value string }{
					*tags[k].Key,
					*tags[k].Value,
				}
				if e == g {
					continue EXPECTED_TAG
				}
			}
			t.Errorf("%d, expected %+v, but not found in output tags", i, e)
		}
	}
}

func TestParseTagsErrors(t *testing.T) {
	cases := []string{
		"bad:\nyaml",
		`{"json":["with","a","list"]}`,
		`["not","key","value","map"]`,
	}

	for i, c := range cases {
		_, err := parseTags(c)
		if err == nil {
			t.Errorf("%d, expected error, but got success", i)
		}
	}
}

func TestParseParameters(t *testing.T) {
	cases := []struct {
		input              []string
		expectedParameters []*cloudformation.Parameter
		envVars            map[string]string
	}{
		// JSON String
		{
			input: []string{`{"ThisKey":"ThisValue"}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("ThisValue"),
				},
			},
		},
		// YAML String
		{
			input: []string{"---\nThisKey: ThisValue"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("ThisValue"),
				},
			},
		},
		// JSON Integer
		{
			input: []string{`{"ThisKey":123}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("123"),
				},
			},
		},
		// YAML Integer
		{
			input: []string{"---\nThisKey: 123"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("123"),
				},
			},
		},
		// JSON Float
		{
			input: []string{`{"ThisKey": 3.1415926535897932384626433832795028841971}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("3.141592653589793"),
				},
			},
		},
		// YAML Float
		{
			input: []string{"---\nThisKey: 3.1415926535897932384626433832795028841971"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("3.141592653589793"),
				},
			},
		},
		// JSON Bool
		{
			input: []string{`{"ThisKey":true}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("true"),
				},
			},
		},
		// YAML Bool
		{
			input: []string{"---\nThisKey: true"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("true"),
				},
			},
		},
		// JSON String List
		{
			input: []string{`{"ThisKey":["one","two"]}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,two"),
				},
			},
		},
		// YAML String List
		{
			input: []string{"---\nThisKey:\n  - one\n  - two"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,two"),
				},
			},
		},
		// JSON Misc-List
		{
			input: []string{`{"ThisKey":["one",2,123.456,false]}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,2,123.456,false"),
				},
			},
		},
		// YAML Misc-List
		{
			input: []string{"---\nThisKey:\n  - one\n  - 2\n  - 123.456\n  - false"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,2,123.456,false"),
				},
			},
		},
		// JSON Multi-value
		{
			input: []string{`{
				"String": "Foobar",
                "Int": 123,
                "Float": 123.456,
				"Boolean": true,
				"List": ["one","two","three"]
			}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("String"),
					ParameterValue: aws.String("Foobar"),
				},
				{
					ParameterKey:   aws.String("Int"),
					ParameterValue: aws.String("123"),
				},
				{
					ParameterKey:   aws.String("Float"),
					ParameterValue: aws.String("123.456"),
				},
				{
					ParameterKey:   aws.String("Boolean"),
					ParameterValue: aws.String("true"),
				},
				{
					ParameterKey:   aws.String("List"),
					ParameterValue: aws.String("one,two,three"),
				},
			},
		},
		// YAML Multi-value
		{
			input: []string{`---
                String: Foobar
                Int: 123
                Float: 123.456
                Boolean: true
                List:
                    - one
                    - two
                    - three`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("String"),
					ParameterValue: aws.String("Foobar"),
				},
				{
					ParameterKey:   aws.String("Int"),
					ParameterValue: aws.String("123"),
				},
				{
					ParameterKey:   aws.String("Float"),
					ParameterValue: aws.String("123.456"),
				},
				{
					ParameterKey:   aws.String("Boolean"),
					ParameterValue: aws.String("true"),
				},
				{
					ParameterKey:   aws.String("List"),
					ParameterValue: aws.String("one,two,three"),
				},
			},
		},
		// JSON String, with env variables
		{
			input: []string{`{"ThisKey":"ThisValue-{{ env \"TEST_VAR1\"}}"}`},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("ThisValue-VALUE_HERE"),
				},
			},
			envVars: map[string]string{"TEST_VAR1": "VALUE_HERE"},
		},
		// YAML String, with env variables
		{
			input: []string{"---\nThisKey: ThisValue-{{ env \"TEST_VAR1\"}}"},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("ThisValue-VALUE_HERE"),
				},
			},
			envVars: map[string]string{"TEST_VAR1": "VALUE_HERE"},
		},
		// JSON String, with parameter overrides (all overrides)
		{
			input: []string{
				`{"One":"Foo","Two":"Foo"}`,
				`{"One":"Bar","Two":"Bar"}`,
			},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("One"),
					ParameterValue: aws.String("Bar"),
				},
				{
					ParameterKey:   aws.String("Two"),
					ParameterValue: aws.String("Bar"),
				},
			},
		},
		// YAML String, with parameter overrides (all overrides)
		{
			input: []string{
				"---\nOne: Foo\nTwo: Foo",
				"---\nOne: Bar\nTwo: Bar",
			},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("One"),
					ParameterValue: aws.String("Bar"),
				},
				{
					ParameterKey:   aws.String("Two"),
					ParameterValue: aws.String("Bar"),
				},
			},
		},
		// JSON String, with parameter overrides (some overrides)
		{
			input: []string{
				`{"One":"Foo","Two":"Foo"}`,
				`{"Two":"Bar"}`,
			},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("One"),
					ParameterValue: aws.String("Foo"),
				},
				{
					ParameterKey:   aws.String("Two"),
					ParameterValue: aws.String("Bar"),
				},
			},
		},
		// YAML String, with parameter overrides (some overrides)
		{
			input: []string{
				"---\nOne: Foo\nTwo: Foo",
				"---\nTwo: Bar",
			},
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("One"),
					ParameterValue: aws.String("Foo"),
				},
				{
					ParameterKey:   aws.String("Two"),
					ParameterValue: aws.String("Bar"),
				},
			},
		},
	}

	for i, c := range cases {
		oldValueMap := map[string]string{}
		for k, v := range c.envVars {
			if oldValue, present := os.LookupEnv(k); present {
				oldValueMap[k] = oldValue
				defer os.Setenv(k, oldValueMap[k])
			}
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}

		var parameters []*cloudformation.Parameter
		parameters, err := parseParameters(c.input)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

		for j := 0; j < len(c.expectedParameters); j++ {
			found := false
			e := struct{ ParameterKey, ParameterValue string }{
				*c.expectedParameters[j].ParameterKey,
				*c.expectedParameters[j].ParameterValue,
			}
			for k := 0; k < len(parameters); k++ {
				g := struct{ ParameterKey, ParameterValue string }{
					*parameters[k].ParameterKey,
					*parameters[k].ParameterValue,
				}
				if e.ParameterKey == g.ParameterKey {
					if e.ParameterValue == g.ParameterValue {
						found = true
					} else {
						t.Errorf("%d, expected %+v, got %+v", i, e, g)
					}
				}
			}
			if !found {
				t.Errorf("%d, expected %+v, but not found in output parameters", i, e)
			}
		}
	}
}

func TestParseParametersErrors(t *testing.T) {
	cases := []string{
		"bad:\nyaml",
		`{"nested":{"objects":"here"}}`,
		`["not","key","value","map"]`,
		`{"Sublist":[["one","two"],"three"]}`,
		`{"CommaInList":["foo,bar","foobar"]}`,
	}

	for i, c := range cases {
		_, err := parseParameters([]string{c})
		if err == nil {
			t.Errorf("%d, expected error, but got success", i)
		}
	}
}

func TestParseEnvironmentVariables(t *testing.T) {
	cases := []struct {
		envVars        map[string]string
		inputTemplate  string
		expectedOutput string
	}{
		{
			envVars: map[string]string{
				"TEST_VAR3": "soup",
				"TEST_VAR4": "TEST=VALUE4",
			},
			inputTemplate:  "This {{ env \"TEST_VAR3\" }} should be good",
			expectedOutput: "This soup should be good",
		},
		{
			envVars: map[string]string{
				"TEST_VAR3": "soup",
				"TEST_VAR4": "TEST=VALUE4",
			},
			inputTemplate:  "This {{env \"TEST_VAR3\"}} should be good. Also, {{env \"TEST_VAR4\"}}",
			expectedOutput: "This soup should be good. Also, TEST=VALUE4",
		},
	}

	for i, c := range cases {
		oldValueMap := map[string]string{}
		for k, v := range c.envVars {
			if oldValue, present := os.LookupEnv(k); present {
				oldValueMap[k] = oldValue
				defer os.Setenv(k, oldValueMap[k])
			}
			os.Setenv(k, v)
			defer os.Unsetenv(k)
		}

		parsedInput, err := parseEnvironmentVariables(c.inputTemplate)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if e, g := c.expectedOutput, parsedInput; e != g {
			t.Errorf("%d, expected \"%s\", got \"%s\"", i, e, g)
		}
	}
}

func TestParseEnvironmentVariablesError(t *testing.T) {
	cases := []struct {
		envVars       []string
		inputTemplate string
	}{
		{
			envVars:       []string{"SHOULD_NOT_EXIST"},
			inputTemplate: "This {{env \"SHOULD_NOT_EXIST\"}} should fail because the env var isn't defined",
		},
	}

	for i, c := range cases {
		oldValueMap := map[string]string{}
		for _, k := range c.envVars {
			if oldValue, present := os.LookupEnv(k); present {
				oldValueMap[k] = oldValue
				defer os.Setenv(k, oldValueMap[k])
			}
			os.Unsetenv(k)
		}

		_, err := parseEnvironmentVariables(c.inputTemplate)
		if err == nil {
			t.Errorf("%d, expected error, but got success", i)
		}
	}
}
