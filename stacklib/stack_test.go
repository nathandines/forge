package stacklib

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type mockStacks struct {
	stacksOutput cloudformation.DescribeStacksOutput
	cloudformationiface.CloudFormationAPI
}

func (m mockStacks) DescribeStacks(*cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	return &m.stacksOutput, nil
}

func TestGetStackInfo(t *testing.T) {
	cases := []struct {
		stackName string
		stackID   string
		resp      cloudformation.DescribeStacksOutput
		expect    *cloudformation.Stack
	}{
		{
			stackName: "test-stack",
			resp: cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					{
						StackName: aws.String("test-stack"),
						StackId:   aws.String("arn:aws:cloudformation:ap-southeast-2:012345678901:stack/test-stack/stackid0"),
					},
				},
			},
		},
		{
			stackID: "arn:aws:cloudformation:ap-southeast-2:012345678901:stack/test-stack/stackid1",
			resp: cloudformation.DescribeStacksOutput{
				Stacks: []*cloudformation.Stack{
					{
						StackName: aws.String("test-stack"),
						StackId:   aws.String("arn:aws:cloudformation:ap-southeast-2:012345678901:stack/test-stack/stackid1"),
					},
				},
			},
		},
	}

	for i, c := range cases {
		cfn = mockStacks{stacksOutput: c.resp}

		s := Stack{
			StackName: c.stackName,
			StackID:   c.stackID,
		}

		err := s.GetStackInfo()
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

		// Test that the expected values are populated based on the
		// CloudFormation response
		if e, g := *c.resp.Stacks[0].StackName, s.StackName; e != g {
			t.Errorf("%d, expected \"%s\" stack name, got \"%s\"", i, e, g)
		}
		if e, g := *c.resp.Stacks[0].StackId, s.StackID; e != g {
			t.Errorf("%d, expected \"%s\" stack id, got \"%s\"", i, e, g)
		}
		if e, g := c.resp.Stacks[0], s.StackInfo; e != g {
			t.Errorf("%d, expected %v stack info, got %v", i, e, g)
		}
	}
}

// TestGetStackInfoNoStackParameters must return an error, as stack info cannot
// be looked up without any information
func TestGetStackInfoNoStackParameters(t *testing.T) {
	s := Stack{}

	if err := s.GetStackInfo(); err == nil {
		t.Errorf("expected error, got success")
	}
}
