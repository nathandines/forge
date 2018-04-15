package forgelib

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type mockDelete struct {
	stacks *[]cloudformation.Stack
	cloudformationiface.CloudFormationAPI
}

func (m mockDelete) DeleteStack(input *cloudformation.DeleteStackInput) (output *cloudformation.DeleteStackOutput, err error) {
	for i := 0; i < len(*m.stacks); i++ {
		if *(*m.stacks)[i].StackId == *input.StackName &&
			*(*m.stacks)[i].StackStatus != cloudformation.StackStatusDeleteComplete {
			*(*m.stacks)[i].StackStatus = cloudformation.StackStatusDeleteComplete
			return
		}
	}
	return output, fmt.Errorf("stack not found")
}

func TestDestroy(t *testing.T) {
	cases := []struct {
		expectStacks  []cloudformation.Stack
		expectSuccess bool
		thisStack     Stack
		stacks        []cloudformation.Stack
	}{
		{
			thisStack: Stack{
				StackID: "test-stack/id0",
			},
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
			},
			expectSuccess: true,
		},
		{
			thisStack: Stack{
				StackID: "test-stack/id0",
			},
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
			},
			expectSuccess: false,
		},
	}

	oldCFNClient := cfnClient
	defer func() { cfnClient = oldCFNClient }()
	for i, c := range cases {
		theseStacks := cases[i].stacks
		cfnClient = mockDelete{
			stacks: &theseStacks,
		}

		err := c.thisStack.Destroy()
		if err != nil && c.expectSuccess {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if err == nil && !c.expectSuccess {
			t.Errorf("%d, expected error, got success", i)
		}

		for j := 0; j < len(c.expectStacks); j++ {
			e := struct{ StackName, StackID, StackStatus string }{
				*c.expectStacks[j].StackName,
				*c.expectStacks[j].StackId,
				*c.expectStacks[j].StackStatus,
			}
			g := struct{ StackName, StackID, StackStatus string }{
				*theseStacks[j].StackName,
				*theseStacks[j].StackId,
				*theseStacks[j].StackStatus,
			}
			if e != g {
				t.Errorf("%d, expected %+v, got %+v", i, e, g)
			}
		}
	}
}

func TestDestroyNoStackID(t *testing.T) {
	s := Stack{}

	if err := s.Destroy(); err == nil {
		t.Errorf("expected error, got success")
	}
}
