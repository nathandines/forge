package stacklib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type mockDeploy struct {
	capabilityIam bool
	failCreate    bool
	failDescribe  bool
	failValidate  bool
	newStackID    string
	noUpdates     bool
	stacks        *[]cloudformation.Stack
	cloudformationiface.CloudFormationAPI
}

type fakeReadFile struct {
	String string
}

type fakeStack struct {
	StackName, StackID, StackStatus string
	Tags                            []fakeTag
}

type fakeTag struct {
	Key, Value string
}

type byKey []fakeTag

func (t byKey) Len() int           { return len(t) }
func (t byKey) Less(i, j int) bool { return t[i].Key < t[j].Key }
func (t byKey) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func genFakeStackData(realStack cloudformation.Stack) fakeStack {
	output := fakeStack{
		StackName:   *realStack.StackName,
		StackID:     *realStack.StackId,
		StackStatus: *realStack.StackStatus,
	}
	for _, t := range realStack.Tags {
		output.Tags = append(output.Tags, fakeTag{
			Key:   *t.Key,
			Value: *t.Value,
		})
	}
	sort.Sort(byKey(output.Tags))
	return output
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

	if m.failValidate {
		return &output, awserr.New(
			"ValidationError",
			"Invalid template property or properties [BadProperty]",
			nil,
		)
	}
	if m.capabilityIam {
		output.Capabilities = aws.StringSlice([]string{
			cloudformation.CapabilityCapabilityIam,
		})
	}
	return &output, nil
}

func (m mockDeploy) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	output := cloudformation.CreateStackOutput{}

	if m.failCreate {
		return &output, awserr.New(
			cloudformation.ErrCodeInvalidOperationException,
			"Simulated Failure",
			nil,
		)
	}

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
		Tags:        input.Tags,
	}
	*m.stacks = append(*m.stacks, thisStack)
	output.StackId = &m.newStackID
	return &output, nil
}

func (m mockDeploy) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	output := cloudformation.UpdateStackOutput{}
	if m.capabilityIam {
		if err := testIamCapability(input.Capabilities); err != nil {
			return &output, err
		}
	}

	if m.noUpdates {
		return &output, awserr.New(
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
				(*m.stacks)[i].Tags = input.Tags
				output.StackId = &m.newStackID
				return &output, nil
			}
		}
	}

	return &output, awserr.New(
		"ValidationError",
		fmt.Sprintf("Stack with id %s does not exist", *input.StackName),
		nil,
	)
}

func (m mockDeploy) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
	var err error
	output := cloudformation.DescribeStacksOutput{}
	outputStacks := []*cloudformation.Stack{}

	if m.failDescribe {
		return &output, awserr.New(
			cloudformation.ErrCodeInvalidOperationException,
			"Simulated Failure",
			nil,
		)
	}

	for i := 0; i < len(*m.stacks); i++ {
		if s := *input.StackName; s != "" {
			if s == *(*m.stacks)[i].StackId ||
				(s == *(*m.stacks)[i].StackName &&
					*(*m.stacks)[i].StackStatus != cloudformation.StackStatusDeleteComplete) {
				outputStacks = append(outputStacks, &(*m.stacks)[i])
				output.Stacks = outputStacks
				return &output, err
			}
		} else {
			outputStacks = append(outputStacks, &(*m.stacks)[i])
			output.Stacks = outputStacks
			return &output, err
		}
	}
	if *input.StackName != "" {
		err = awserr.New(
			"ValidationError",
			fmt.Sprintf("Stack with id %s does not exist", *input.StackName),
			nil,
		)
	}
	return &output, err
}

func (f fakeReadFile) readFile(filename string) ([]byte, error) {
	buf := bytes.NewBufferString(f.String)
	return ioutil.ReadAll(buf)
}

func TestDeploy(t *testing.T) {
	cases := []struct {
		capabilityIam bool
		failCreate    bool
		failDescribe  bool
		failValidate  bool
		expectFailure bool
		expectOutput  DeployOut
		expectStacks  []cloudformation.Stack
		newStackID    string
		noUpdates     bool
		stacks        []cloudformation.Stack
		tagInput      string
		thisStack     Stack
	}{
		// Create new stack with previously used name
		{
			newStackID: "test-stack/id2",
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
		},
		// Create new stack where one did not previously exist
		{
			newStackID: "test-stack/id0",
			stacks:     []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
		},
		// Update stack where one was previously created
		{
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
		},
		// Test deployment against a non-deployable state
		// This covers what would have otherwise been failUpdate
		{
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
			expectFailure: true,
		},
		// Test successful behaviour when no updates are to be performed
		{
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
			noUpdates:    true,
			expectOutput: DeployOut{Message: "No updates are to be performed."},
		},
		// Require CAPABILITY_IAM Stack Update
		{
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
			capabilityIam: true,
		},
		// Require CAPABILITY_IAM Stack Create
		{
			newStackID: "test-stack/id0",
			stacks:     []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			capabilityIam: true,
		},
		// CreateStack error
		{
			stacks:        []cloudformation.Stack{},
			expectStacks:  []cloudformation.Stack{},
			expectFailure: true,
			failCreate:    true,
		},
		// DescribeStacks error
		{
			stacks:        []cloudformation.Stack{},
			expectStacks:  []cloudformation.Stack{},
			expectFailure: true,
			failDescribe:  true,
		},
		// ValidateTemplate error
		{
			stacks:        []cloudformation.Stack{},
			expectStacks:  []cloudformation.Stack{},
			expectFailure: true,
			failValidate:  true,
		},
		// Create stack with tags
		{
			newStackID: "test-stack/id0",
			tagInput:   `{"TestKey1":"TestValue1","TestKey2":"TestValue2"}`,
			stacks:     []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Tags: []*cloudformation.Tag{
						{Key: aws.String("TestKey1"), Value: aws.String("TestValue1")},
						{Key: aws.String("TestKey2"), Value: aws.String("TestValue2")},
					},
				},
			},
		},
		// Update stack, adding tags
		{
			tagInput: `{"TestKey1":"TestValue1","TestKey2":"TestValue2"}`,
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
					Tags: []*cloudformation.Tag{
						{Key: aws.String("TestKey1"), Value: aws.String("TestValue1")},
						{Key: aws.String("TestKey2"), Value: aws.String("TestValue2")},
					},
				},
			},
		},
	}

	for i, c := range cases {
		theseStacks := cases[i].stacks
		cfn = mockDeploy{
			capabilityIam: c.capabilityIam,
			failCreate:    c.failCreate,
			failDescribe:  c.failDescribe,
			failValidate:  c.failValidate,
			newStackID:    c.newStackID,
			noUpdates:     c.noUpdates,
			stacks:        &theseStacks,
		}

		fakeTemplate := fakeReadFile{String: `{"Resources":{"SNS":{"Type":"AWS::SNS::Topic"}}}`}
		readTemplate = fakeTemplate.readFile

		tagsFile := ""
		if c.tagInput != "" {
			fakeTags := fakeReadFile{String: c.tagInput}
			readTags = fakeTags.readFile
			tagsFile = "whatever_tags.yml"
		}

		thisStack := c.thisStack
		if thisStack == (Stack{}) {
			thisStack = Stack{
				StackName:    "test-stack",
				TagsFile:     tagsFile,
				TemplateFile: "whatever.yml",
			}
		}

		output, err := thisStack.Deploy()
		switch {
		case err == nil && c.expectFailure:
			t.Errorf("%d, expected error, got success", i)
		case err != nil && !c.expectFailure:
			t.Fatalf("%d, unexpected error, %v", i, err)
		}

		if e, g := c.expectOutput, output; e != g {
			t.Errorf("%d, expected %+v info, got %+v", i, e, g)
		}

		for j := 0; j < len(c.expectStacks); j++ {
			e := genFakeStackData(c.expectStacks[j])
			g := genFakeStackData(theseStacks[j])
			if !reflect.DeepEqual(e, g) {
				t.Errorf("%d, expected %+v, got %+v", i, e, g)
			}
		}
	}
}
