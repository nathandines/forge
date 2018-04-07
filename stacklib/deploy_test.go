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
	newStackID    string
	noUpdates     bool
	capabilityIam bool
	stacks        *[]cloudformation.Stack
	cloudformationiface.CloudFormationAPI
}

type fakeReadFile struct {
	String string
}

func testIamCapability(inputCapabilities []*string) (err error) {
	for _, c := range inputCapabilities {
		if *c == cloudformation.CapabilityCapabilityIam {
			return err
		}
	}
	return awserr.New(
		cloudformation.ErrCodeInsufficientCapabilitiesException,
		"Requires capabilities : [CAPABILITY_IAM]",
		nil,
	)
}

func (m mockDeploy) ValidateTemplate(*cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
	output := cloudformation.ValidateTemplateOutput{}
	if m.capabilityIam {
		output.Capabilities = aws.StringSlice([]string{
			cloudformation.CapabilityCapabilityIam,
		})
	}
	return &output, nil
}

func (m mockDeploy) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	output := cloudformation.CreateStackOutput{}
	if m.capabilityIam {
		if err := testIamCapability(input.Capabilities); err != nil {
			return &output, err
		}
	}
	for i := 0; i < len(*m.stacks); i++ {
		if *(*m.stacks)[i].StackName == *input.StackName &&
			*(*m.stacks)[i].StackStatus != cloudformation.StackStatusDeleteComplete {
			return &output, awserr.New(
				cloudformation.ErrCodeAlreadyExistsException,
				fmt.Sprintf("Stack [%s] already exists", *input.StackName),
				nil,
			)
		}
	}
	thisStack := cloudformation.Stack{
		StackName:   input.StackName,
		StackId:     aws.String(m.newStackID),
		StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
	}
	*m.stacks = append(*m.stacks, thisStack)
	output.StackId = &m.newStackID
	return &output, nil
}

func (m mockDeploy) UpdateStack(input *cloudformation.UpdateStackInput) (output *cloudformation.UpdateStackOutput, err error) {
	if m.capabilityIam {
		if err := testIamCapability(input.Capabilities); err != nil {
			return output, err
		}
	}
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
	return output, awserr.New(
		"ValidationError",
		fmt.Sprintf("Stack with id %s does not exist", *input.StackName),
		nil,
	)
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
		expectOutput  DeployOut
		expectStacks  []cloudformation.Stack
		expectSuccess bool
		newStackID    string
		noUpdates     bool
		capabilityIam bool
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
			newStackID: "test-stack/id0",
			thisStack: Stack{
				StackName: "test-stack",
			},
			stacks: []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
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
			expectOutput:  DeployOut{Message: "No updates are to be performed."},
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
			capabilityIam: true,
		},
		{
			newStackID: "test-stack/id0",
			thisStack: Stack{
				StackName: "test-stack",
			},
			stacks: []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectSuccess: true,
			capabilityIam: true,
		},
	}

	for i, c := range cases {
		theseStacks := cases[i].stacks
		cfn = mockDeploy{
			stacks:        &theseStacks,
			newStackID:    c.newStackID,
			noUpdates:     c.noUpdates,
			capabilityIam: c.capabilityIam,
		}

		fakeIO := fakeReadFile{String: `{"Resources":{"SNS":{"Type":"AWS::SNS::Topic"}}}`}
		readFile = fakeIO.readFile

		c.thisStack.TemplateFile = "whatever.yml"

		output, err := c.thisStack.Deploy()
		if err != nil && c.expectSuccess {
			t.Fatalf("%d, unexpected error, %v", i, err)
		}
		if err == nil && !c.expectSuccess {
			t.Errorf("%d, expected error, got success", i)
		}
		if e, g := c.expectOutput, output; e != g {
			t.Errorf("%d, expected %+v info, got %+v", i, e, g)
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
