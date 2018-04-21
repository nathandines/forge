package forgelib

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

type fakeStack struct {
	RoleARN, StackName, StackID, StackStatus string
	Tags                                     []fakeTag
	Parameters                               []fakeParameter
}

type fakeTag struct {
	Key, Value string
}

type fakeParameter struct {
	ParameterKey, ParameterValue string
}

type byKey []fakeTag

func (t byKey) Len() int           { return len(t) }
func (t byKey) Less(i, j int) bool { return t[i].Key < t[j].Key }
func (t byKey) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

type byParameterKey []fakeParameter

func (t byParameterKey) Len() int           { return len(t) }
func (t byParameterKey) Less(i, j int) bool { return t[i].ParameterKey < t[j].ParameterKey }
func (t byParameterKey) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

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
	for _, p := range realStack.Parameters {
		output.Parameters = append(output.Parameters, fakeParameter{
			ParameterKey:   *p.ParameterKey,
			ParameterValue: *p.ParameterValue,
		})
	}
	sort.Sort(byParameterKey(output.Parameters))

	if r := realStack.RoleARN; r != nil {
		output.RoleARN = *r
	}

	return output
}

func TestDeploy(t *testing.T) {
	cases := []struct {
		accountID          string
		capabilityIam      bool
		cfnRoleName        string
		expectFailure      bool
		expectOutput       DeployOut
		expectStacks       []cloudformation.Stack
		expectStackPolicy  string
		failCreate         bool
		failDescribe       bool
		failValidate       bool
		newStackID         string
		noUpdates          bool
		parameterInput     string
		requiredParameters []string
		stacks             []cloudformation.Stack
		stackPolicies      map[string]string
		stackPolicyInput   string
		tagInput           string
		thisStack          Stack
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
		// Update stack without tags file, don't remove tags
		{
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Tags: []*cloudformation.Tag{
						{Key: aws.String("TestKey1"), Value: aws.String("TestValue1")},
						{Key: aws.String("TestKey2"), Value: aws.String("TestValue2")},
					},
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
		// Update stack, remove tags
		{
			tagInput: "{}",
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Tags: []*cloudformation.Tag{
						{Key: aws.String("TestKey1"), Value: aws.String("TestValue1")},
						{Key: aws.String("TestKey2"), Value: aws.String("TestValue2")},
					},
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
				},
			},
		},
		// Create stack with parameters
		{
			newStackID:         "test-stack/id0",
			parameterInput:     `{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`,
			requiredParameters: []string{"TestParam1", "TestParam2"},
			stacks:             []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("TestParam1"), ParameterValue: aws.String("TestValue1")},
						{ParameterKey: aws.String("TestParam2"), ParameterValue: aws.String("TestValue2")},
					},
				},
			},
		},
		// Update stack, adding parameters
		{
			parameterInput:     `{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`,
			requiredParameters: []string{"TestParam1", "TestParam2"},
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
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("TestParam1"), ParameterValue: aws.String("TestValue1")},
						{ParameterKey: aws.String("TestParam2"), ParameterValue: aws.String("TestValue2")},
					},
				},
			},
		},
		// Update stack with subset of parameters
		{
			parameterInput:     `{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`,
			requiredParameters: []string{"TestParam1"},
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("TestParam1"), ParameterValue: aws.String("TestValue1")},
						{ParameterKey: aws.String("TestParam2"), ParameterValue: aws.String("TestValue2")},
					},
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id1"),
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("TestParam1"), ParameterValue: aws.String("TestValue1")},
					},
				},
			},
		},
		// Create stack with subset of parameters
		{
			newStackID:         "test-stack/id0",
			parameterInput:     `{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`,
			requiredParameters: []string{"TestParam1"},
			stacks:             []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("TestParam1"), ParameterValue: aws.String("TestValue1")},
					},
				},
			},
		},
		// Create stack, missing required parameters
		{
			newStackID:         "test-stack/id0",
			parameterInput:     `{"TestParam2":"TestValue2"}`,
			requiredParameters: []string{"TestParam1"},
			stacks:             []cloudformation.Stack{},
			expectStacks:       []cloudformation.Stack{},
			expectFailure:      true,
		},
		// Update stack, missing required parameters
		{
			parameterInput:     `{"TestParam2":"TestValue2"}`,
			requiredParameters: []string{"TestParam1"},
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
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectFailure: true,
		},
		// Create stack, using role
		{
			cfnRoleName: "role-name",
			accountID:   "111111111111",
			newStackID:  "test-stack/id0",
			stacks:      []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					RoleARN:     aws.String("arn:aws:iam::111111111111:role/role-name"),
				},
			},
		},
		// Update stack, using role
		{
			cfnRoleName: "role-name",
			accountID:   "111111111111",
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
					RoleARN:     aws.String("arn:aws:iam::111111111111:role/role-name"),
				},
			},
		},
		// Create; JSON Stack Policy
		{
			newStackID: "test-stack/id0",
			stackPolicyInput: `{
				"Statement":
				[
					{
						"Effect": "Allow",
						"Action" : "Update:*",
						"Principal": "*",
						"NotResource" : "LogicalResourceId/ProductionDatabase"
					}
				]
			}`,
			stacks: []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
		// Update; JSON Stack Policy
		{
			stackPolicyInput: `{
				"Statement":
				[
					{
						"Effect": "Allow",
						"Action" : "Update:*",
						"Principal": "*",
						"NotResource" : "LogicalResourceId/ProductionDatabase"
					}
				]
			}`,
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
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
		// Create; YAML Stack Policy
		{
			newStackID: "test-stack/id0",
			stackPolicyInput: `---
                Statement:
                - Effect: Allow
                  Action: Update:*
                  Principal: '*'
                  NotResource: LogicalResourceId/ProductionDatabase`,
			stacks: []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
		// Update; YAML Stack Policy
		{
			stackPolicyInput: `---
                Statement:
                - Effect: Allow
                  Action: Update:*
                  Principal: '*'
                  NotResource: LogicalResourceId/ProductionDatabase`,
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
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
		// Update; Existing JSON Stack Policy without update
		{
			stackPolicyInput: "",
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
			stackPolicies: map[string]string{
				"test-stack/id1": `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
			},
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
	}

	oldCFNClient := cfnClient
	defer func() { cfnClient = oldCFNClient }()
	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()
	for i, c := range cases {
		theseStacks := cases[i].stacks
		theseStackPolicies := c.stackPolicies
		if theseStackPolicies == nil {
			theseStackPolicies = map[string]string{}
		}
		cfnClient = mockCfn{
			capabilityIam:      c.capabilityIam,
			failCreate:         c.failCreate,
			failDescribe:       c.failDescribe,
			failValidate:       c.failValidate,
			newStackID:         c.newStackID,
			noUpdates:          c.noUpdates,
			requiredParameters: c.requiredParameters,
			stacks:             &theseStacks,
			stackPolicies:      &theseStackPolicies,
		}
		stsClient = mockSTS{accountID: c.accountID}

		thisStack := c.thisStack
		if thisStack == (Stack{}) {
			thisStack = Stack{
				ParametersBody:  c.parameterInput,
				StackName:       "test-stack",
				TagsBody:        c.tagInput,
				TemplateBody:    `{"Resources":{"SNS":{"Type":"AWS::SNS::Topic"}}}`,
				CfnRoleName:     c.cfnRoleName,
				StackPolicyBody: c.stackPolicyInput,
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

		if thisStack.StackID == "" && !c.expectFailure {
			t.Errorf("%d, expected populated Stack ID on stack, found none", i)
		}

		for j := 0; j < len(c.expectStacks); j++ {
			e := genFakeStackData(c.expectStacks[j])
			g := genFakeStackData(theseStacks[j])
			if !reflect.DeepEqual(e, g) {
				t.Errorf("%d, expected %+v, got %+v", i, e, g)
			}
		}

		if e := c.expectStackPolicy; e != "" {
			if g, ok := theseStackPolicies[thisStack.StackID]; ok {
				if e != g {
					t.Errorf("%d, expected stack policy \"%s\", got \"%s\"", i, e, g)
				}
			} else {
				t.Errorf("%d, expected stack policy \"%s\", got none", i, e)
			}
		}
	}
}
