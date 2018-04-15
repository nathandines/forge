package forgelib

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestParseTags(t *testing.T) {
	cases := []struct {
		input        string
		expectedTags []*cloudformation.Tag
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
	}

	for i, c := range cases {
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
		input              string
		expectedParameters []*cloudformation.Parameter
	}{
		// JSON String
		{
			input: `{"ThisKey":"ThisValue"}`,
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("ThisValue"),
				},
			},
		},
		// YAML String
		{
			input: "---\nThisKey: ThisValue",
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("ThisValue"),
				},
			},
		},
		// JSON Integer
		{
			input: `{"ThisKey":123}`,
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("123"),
				},
			},
		},
		// YAML Integer
		{
			input: "---\nThisKey: 123",
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("123"),
				},
			},
		},
		// JSON Float
		{
			input: `{"ThisKey": 3.1415926535897932384626433832795028841971}`,
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("3.141592653589793"),
				},
			},
		},
		// YAML Float
		{
			input: "---\nThisKey: 3.1415926535897932384626433832795028841971",
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("3.141592653589793"),
				},
			},
		},
		// JSON Bool
		{
			input: `{"ThisKey":true}`,
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("true"),
				},
			},
		},
		// YAML Bool
		{
			input: "---\nThisKey: true",
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("true"),
				},
			},
		},
		// JSON String List
		{
			input: `{"ThisKey":["one","two"]}`,
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,two"),
				},
			},
		},
		// YAML String List
		{
			input: "---\nThisKey:\n  - one\n  - two",
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,two"),
				},
			},
		},
		// JSON Misc-List
		{
			input: `{"ThisKey":["one",2,123.456,false]}`,
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,2,123.456,false"),
				},
			},
		},
		// YAML Misc-List
		{
			input: "---\nThisKey:\n  - one\n  - 2\n  - 123.456\n  - false",
			expectedParameters: []*cloudformation.Parameter{
				{
					ParameterKey:   aws.String("ThisKey"),
					ParameterValue: aws.String("one,2,123.456,false"),
				},
			},
		},
		// JSON Multi-value
		{
			input: `{
				"String": "Foobar",
                "Int": 123,
                "Float": 123.456,
				"Boolean": true,
				"List": ["one","two","three"]
			}`,
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
			input: `---
                String: Foobar
                Int: 123
                Float: 123.456
                Boolean: true
                List:
                    - one
                    - two
                    - three`,
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
	}

	for i, c := range cases {
		var parameters []*cloudformation.Parameter
		parameters, err := parseParameters(c.input)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

	EXPECTED_PARAMETER:
		for j := 0; j < len(c.expectedParameters); j++ {
			e := struct{ ParameterKey, ParameterValue string }{
				*c.expectedParameters[j].ParameterKey,
				*c.expectedParameters[j].ParameterValue,
			}
			for k := 0; k < len(parameters); k++ {
				g := struct{ ParameterKey, ParameterValue string }{
					*parameters[k].ParameterKey,
					*parameters[k].ParameterValue,
				}
				if e == g {
					continue EXPECTED_PARAMETER
				}
			}
			t.Errorf("%d, expected %+v, but not found in output parameters", i, e)
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
		_, err := parseParameters(c)
		if err == nil {
			t.Errorf("%d, expected error, but got success", i)
		}
	}
}
