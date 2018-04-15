package forgelib

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

func TestDestroy(t *testing.T) {
	cases := []struct {
		accountID    string
		expectStacks []cloudformation.Stack
		stacks       []cloudformation.Stack
		thisStack    Stack
	}{
		{
			thisStack: Stack{StackID: "test-stack/id0"},
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
		},
		{
			thisStack: Stack{
				StackID:     "test-stack/id0",
				CfnRoleName: "destroy-role-name",
			},
			accountID: "121212121212",
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
					RoleARN:     aws.String("arn:aws:iam::121212121212:role/destroy-role-name"),
				},
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
			},
		},
	}

	oldCFNClient := cfnClient
	defer func() { cfnClient = oldCFNClient }()
	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()
	for i, c := range cases {
		theseStacks := cases[i].stacks
		cfnClient = mockCfn{stacks: &theseStacks}
		stsClient = mockSTS{accountID: c.accountID}

		err := c.thisStack.Destroy()
		if err != nil {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

		for j := 0; j < len(c.expectStacks); j++ {
			e := struct{ RoleARN, StackName, StackID, StackStatus string }{
				"",
				*c.expectStacks[j].StackName,
				*c.expectStacks[j].StackId,
				*c.expectStacks[j].StackStatus,
			}
			if r := c.expectStacks[j].RoleARN; r != nil {
				e.RoleARN = *r
			}
			g := struct{ RoleARN, StackName, StackID, StackStatus string }{
				"",
				*theseStacks[j].StackName,
				*theseStacks[j].StackId,
				*theseStacks[j].StackStatus,
			}
			if r := theseStacks[j].RoleARN; r != nil {
				g.RoleARN = *r
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
