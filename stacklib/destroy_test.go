package stacklib

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type fakeStack struct {
	stackName   string
	stackID     string
	stackStatus string
}

type mockDelete struct {
	stacks *[]fakeStack
	cloudformationiface.CloudFormationAPI
}

func (m mockDelete) DeleteStack(input *cloudformation.DeleteStackInput) (output *cloudformation.DeleteStackOutput, err error) {
	for i := 0; i < len(*m.stacks); i++ {
		if (*m.stacks)[i].stackID == *input.StackName &&
			(*m.stacks)[i].stackStatus != cloudformation.StackStatusDeleteComplete {
			(*m.stacks)[i].stackStatus = cloudformation.StackStatusDeleteComplete
			return
		}
	}
	return output, fmt.Errorf("stack not found")
}

func TestDestroy(t *testing.T) {
	cases := []struct {
		expectStacks  []fakeStack
		expectSuccess bool
		thisStack     Stack
		stacks        []fakeStack
	}{
		{
			thisStack: Stack{
				StackID: "test-stack/id0",
			},
			stacks: []fakeStack{
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id0",
					stackStatus: cloudformation.StackStatusCreateComplete,
				},
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id1",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
			},
			expectStacks: []fakeStack{
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id0",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id1",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
			},
			expectSuccess: true,
		},
		{
			thisStack: Stack{
				StackID: "test-stack/id0",
			},
			stacks: []fakeStack{
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id0",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id1",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
			},
			expectStacks: []fakeStack{
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id0",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
				{
					stackName:   "test-stack",
					stackID:     "test-stack/id1",
					stackStatus: cloudformation.StackStatusDeleteComplete,
				},
			},
			expectSuccess: false,
		},
	}

	for i, c := range cases {
		cfn = mockDelete{
			stacks: &cases[i].stacks,
		}

		err := c.thisStack.Destroy()
		if err != nil && c.expectSuccess {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if err == nil && !c.expectSuccess {
			t.Errorf("%d, expected error, got success", i)
		}

		for j := range c.stacks {
			if e, g := c.expectStacks[j], c.stacks[j]; e != g {
				t.Errorf("%d, expected %v stack name, got %v", i, e, g)
			}
		}
	}
}
