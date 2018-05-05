package commands

import (
	"reflect"
	"testing"
)

func TestParseParameterOverrideArgs(t *testing.T) {
	cases := []struct {
		input          []string
		expectedOutput map[string]string
	}{
		{
			input: []string{
				"One=Foo",
				"Two=Foo=Bar",
			},
			expectedOutput: map[string]string{
				"One": "Foo",
				"Two": "Foo=Bar",
			},
		},
		{
			input: []string{
				"One=Foo",
				"Two=Foo=Bar",
				"One=Bar",
			},
			expectedOutput: map[string]string{
				"One": "Bar",
				"Two": "Foo=Bar",
			},
		},
	}

	for i, c := range cases {
		output, err := parseParameterOverrideArgs(c.input)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

		if !reflect.DeepEqual(c.expectedOutput, output) {
			t.Errorf("%d, expected %+v, got %+v", i, c.expectedOutput, output)
		}
	}
}

func TestParseParameterOverrideArgsError(t *testing.T) {
	cases := []struct {
		input          []string
		expectedOutput map[string]string
	}{
		{
			input: []string{"Two"},
		},
	}

	for i, c := range cases {
		_, err := parseParameterOverrideArgs(c.input)
		if err == nil {
			t.Errorf("%d, expected error, but got success", i)
		}
	}
}
