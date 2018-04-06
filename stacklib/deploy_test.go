package stacklib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type mockDeploy struct {
	newStackID string
	noUpdates  bool
	stacks     *[]cloudformation.Stack
	cloudformationiface.CloudFormationAPI
}

type fakeReadFile struct {
	String string
}

func (m mockDeploy) ValidateTemplate(*cloudformation.ValidateTemplateInput) (output *cloudformation.ValidateTemplateOutput, err error) {
	return
}

func (m mockDeploy) CreateStack(input *cloudformation.CreateStackInput) (output *cloudformation.CreateStackOutput, err error) {
	for i := 0; i < len(*m.stacks); i++ {
		if *(*m.stacks)[i].StackName == *input.StackName &&
			*(*m.stacks)[i].StackStatus != cloudformation.StackStatusDeleteComplete {
			return output, fmt.Errorf("stack already exists")
		}
	}
	thisStack := cloudformation.Stack{
		StackName:   input.StackName,
		StackId:     aws.String(m.newStackID),
		StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
	}
	*m.stacks = append(*m.stacks, thisStack)
	output = &cloudformation.CreateStackOutput{StackId: &m.newStackID}
	return
}

func (m mockDeploy) UpdateStack(input *cloudformation.UpdateStackInput) (output *cloudformation.UpdateStackOutput, err error) {
	if m.noUpdates {
		return output, awserr.New(
			"ValidationError",
			"No updates are to be performed.",
			nil,
		)
	}
	for i := 0; i < len(*m.stacks); i++ {
		if s := *input.StackName; s == *(*m.stacks)[i].StackId ||
			s == *(*m.stacks)[i].StackName {
			switch *(*m.stacks)[i].StackStatus {
			case cloudformation.StackStatusCreateComplete,
				cloudformation.StackStatusUpdateComplete,
				cloudformation.StackStatusUpdateRollbackComplete:
				*(*m.stacks)[i].StackStatus = cloudformation.StackStatusUpdateComplete
				output = &cloudformation.UpdateStackOutput{StackId: &m.newStackID}
				return
			}
		}
	}
	return output, fmt.Errorf("stack in expected state not found")
}

func (m mockDeploy) DescribeStacks(input *cloudformation.DescribeStacksInput) (output *cloudformation.DescribeStacksOutput, err error) {
	outputStacks := []*cloudformation.Stack{}
	for i := 0; i < len(*m.stacks); i++ {
		if s := *input.StackName; s != "" {
			if s == *(*m.stacks)[i].StackId ||
				(s == *(*m.stacks)[i].StackName &&
					*(*m.stacks)[i].StackStatus != cloudformation.StackStatusDeleteComplete) {
				outputStacks = append(outputStacks, &(*m.stacks)[i])
				output = &cloudformation.DescribeStacksOutput{Stacks: outputStacks}
				return
			}
		} else {
			outputStacks = append(outputStacks, &(*m.stacks)[i])
			output = &cloudformation.DescribeStacksOutput{Stacks: outputStacks}
			return
		}
	}
	if *input.StackName != "" {
		err = awserr.New(
			"ValidationError",
			fmt.Sprintf("Stack with id %s does not exist", *input.StackName),
			nil,
		)
	}
	return
}

func (f fakeReadFile) readFile(filename string) ([]byte, error) {
	buf := bytes.NewBufferString(f.String)
	return ioutil.ReadAll(buf)
}

func TestDeploy(t *testing.T) {
	cases := []struct {
		expectStacks  []cloudformation.Stack
		expectSuccess bool
		newStackID    string
		noUpdates     bool
		stacks        []cloudformation.Stack
		thisStack     Stack
	}{
		{
			newStackID: "test-stack/id2",
			thisStack: Stack{
				StackName: "test-stack",
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
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id2"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectSuccess: true,
		},
		{
			thisStack: Stack{
				StackName: "test-stack",
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
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
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
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
				},
			},
			expectSuccess: true,
		},
		{
			thisStack: Stack{
				StackName: "test-stack",
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
					StackStatus: aws.String(cloudformation.StackStatusUpdateRollbackFailed),
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
					StackStatus: aws.String(cloudformation.StackStatusUpdateRollbackFailed),
				},
			},
			expectSuccess: false,
		},
		{
			thisStack: Stack{
				StackName: "test-stack",
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
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
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
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
				},
			},
			expectSuccess: true,
			noUpdates:     true,
		},
	}

	for i, c := range cases {
		theseStacks := cases[i].stacks
		cfn = mockDeploy{
			stacks:     &theseStacks,
			newStackID: c.newStackID,
			noUpdates:  c.noUpdates,
		}

		fakeIO := fakeReadFile{
			String: `{"Resources":{"SNS":{"Type":"AWS::SNS::Topic"}}}`,
		}
		readFile = fakeIO.readFile

		c.thisStack.TemplateFile = "whatever.yml"

		err := c.thisStack.Deploy()
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
