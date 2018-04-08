package stacklib

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
		{
			input: `{"ThisKey":"ThisValue"}`,
			expectedTags: []*cloudformation.Tag{
				{
					Key:   aws.String("ThisKey"),
					Value: aws.String("ThisValue"),
				},
			},
		},
	}

	for i, c := range cases {
		tags, err := parseTags(c.input)
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

		for j := 0; j < len(c.expectedTags); j++ {
			e := struct{ Key, Value string }{
				*c.expectedTags[j].Key,
				*c.expectedTags[j].Value,
			}
			g := struct{ Key, Value string }{
				*tags[j].Key,
				*tags[j].Value,
			}
			if e != g {
				t.Errorf("%d, expected %+v, got %+v", i, e, g)
			}
		}
	}
}
