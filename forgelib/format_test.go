package forgelib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
	"testing"
)

type mockSTS struct {
	accountID string
	stsiface.STSAPI
}

func (m mockSTS) GetCallerIdentity(*sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	output := sts.GetCallerIdentityOutput{
		Account: aws.String(m.accountID),
	}
	return &output, nil
}

func TestValueToStringDefault(t *testing.T) {
	cases := []struct {
		input       interface{}
		expected    string
		allowCommas bool
		allowSlices bool
	}{
		{
			input:    "foobar",
			expected: "foobar",
		},
		// Converting int to float64, as this is how "github.com/ghodss/yaml"
		// behaves
		{
			input:    float64(123),
			expected: "123",
		},
		{
			input:    123.456,
			expected: "123.456",
		},
		{
			input:    true,
			expected: "true",
		},
		{
			input:       "one,two",
			expected:    "one,two",
			allowCommas: true,
		},
		{
			input:       []interface{}{"one", 2.123, true, "foobar"},
			expected:    "one,2.123,true,foobar",
			allowSlices: true,
		},
	}

	for i, c := range cases {
		var output string
		if err := valueToString(c.input, &output, c.allowSlices, c.allowCommas); err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if e, g := c.expected, output; e != g {
			t.Errorf("%d, expected \"%s\", got \"%s\"", i, e, g)
		}
	}
}

func TestValueToStringErrors(t *testing.T) {
	cases := []struct {
		input       interface{}
		allowCommas bool
		allowSlices bool
	}{
		// Fail on slice with comma in a value; allow commas SHOULD NOT apply to
		// values within a slice
		{
			input:       []interface{}{"one", 2.123, true, "foo,bar"},
			allowSlices: true,
			allowCommas: true,
		},
		{
			input:       []interface{}{"one", "two"},
			allowSlices: false,
		},
		{
			input:       "foo,bar",
			allowCommas: false,
		},
	}

	for i, c := range cases {
		var output string
		if err := valueToString(c.input, &output, c.allowSlices, c.allowCommas); err == nil {
			t.Errorf("%d, expected error, but got success", i)
		}
	}
}

func TestRoleARNFromName(t *testing.T) {
	cases := []struct {
		input     string
		accountID string
		expected  string
	}{
		{
			input:     "role-name-here",
			accountID: "123456789012",
			expected:  "arn:aws:iam::123456789012:role/role-name-here",
		},
	}

	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()

	for i, c := range cases {
		stsClient = mockSTS{accountID: c.accountID}

		got, err := roleARNFromName(c.input)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if e, g := c.expected, got; e != g {
			t.Errorf("%d, expected \"%s\", got \"%s\"", i, e, g)
		}
	}
}
