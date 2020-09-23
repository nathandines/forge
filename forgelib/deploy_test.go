package forgelib

import (
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/assert"
)

type fakeStack struct {
	RoleARN, StackName, StackID, StackStatus string
	Tags                                     []fakeTag
	Parameters                               []fakeParameter
	EnableTerminationProtection              bool
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

	if t := realStack.EnableTerminationProtection; t != nil {
		output.EnableTerminationProtection = *t
	}

	return output
}

type resourceValues struct {
	LogicalResourceId string
	ResourceType      string
}

func TestRecursiveSetStackPolicy(t *testing.T) {
	cases := []struct {
		stackPolicyBody string
		stackResources  map[string][]resourceValues
		stackName       string
		fail            bool
	}{
		{
			`{"Statement":[{"Effect":"Allow","Action":["Update:*"],"Principal":"*","Resource":"*"},{"Effect":"Deny","Action":"Update:*","Principal":"*","Resource":"LogicalResourceId/ProductionDatabase"}]}`,
			map[string][]resourceValues{"test-stack": []resourceValues{
				{
					LogicalResourceId: "ProductionDatabase",
					ResourceType:      "AWS::RDS::DBInstance",
				},
			}},
			"test-stack",
			false,
		},
		{
			`{"Statement":[{"Effect":"Allow","Action":["Update:*"],"Principal":"*","Resource":"*"},{"Effect":"Deny","Action":"Update:*","Principal":"*","NotResource":"LogicalResourceId/ProductionDatabase"}]}`,
			map[string][]resourceValues{"test-stack": []resourceValues{
				{
					LogicalResourceId: "ProductionDatabase",
					ResourceType:      "AWS::RDS::DBInstance",
				},
			}},
			"test-stack",
			false,
		},
	}

	for _, c := range cases {

		stackResourcesOutput := make(map[string]cloudformation.DescribeStackResourcesOutput)

		for stackName, stackResources := range c.stackResources {

			var cfStackResources []*cloudformation.StackResource
			for _, stackResource := range stackResources {
				cfStackResource := cloudformation.StackResource{
					LogicalResourceId: &stackResource.LogicalResourceId,
					ResourceType:      &stackResource.ResourceType,
				}
				cfStackResources = append(cfStackResources, &cfStackResource)
			}

			stackResourcesOutput[stackName] = cloudformation.DescribeStackResourcesOutput{
				StackResources: cfStackResources,
			}

		}

		theseStackPolicies := map[string]string{}
		cfnClient = mockCfn{
			stackResourcesOutput: stackResourcesOutput,
			stackPolicies:        &theseStackPolicies,
			newStackID:           c.stackName,
		}

		err := recursiveSetStackPolicy(&c.stackPolicyBody, &c.stackName)
		if err != nil && c.fail == false {
			t.Error(err)
		}
	}
}

func TestDeploy(t *testing.T) {
	cases := []struct {
		testName              string
		accountID             string
		capabilityIam         bool
		cfnRoleName           string
		expectFailure         bool
		expectOutput          DeployOut
		expectStacks          []cloudformation.Stack
		expectStackPolicy     string
		failCreate            bool
		failDescribe          bool
		failValidate          bool
		newStackID            string
		noUpdates             bool
		parameterInput        []string
		parameterOverrides    map[string]string
		requiredParameters    []string
		stacks                []cloudformation.Stack
		stackPolicies         map[string]string
		stackPolicyInput      string
		stackResources        map[string][]resourceValues
		tagInput              string
		terminationProtection bool
	}{
		{
			testName:   "Create new stack with previously used name",
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
		{
			testName:   "Create new stack where one did not previously exist",
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
		{
			testName: "Update stack where one was previously created",
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
			testName: "deployment against a non-deployable state",
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
		{
			testName: "Test successful behaviour when no updates are to be performed",
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
		{
			testName: "Require CAPABILITY_IAM Stack Update",
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
		{
			testName:   "Require CAPABILITY_IAM Stack Create",
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
		{
			testName:      "CreateStack error",
			stacks:        []cloudformation.Stack{},
			expectStacks:  []cloudformation.Stack{},
			expectFailure: true,
			failCreate:    true,
		},
		{
			testName:      "DescribeStacks error",
			stacks:        []cloudformation.Stack{},
			expectStacks:  []cloudformation.Stack{},
			expectFailure: true,
			failDescribe:  true,
		},
		{
			testName:      "ValidateTemplate error",
			stacks:        []cloudformation.Stack{},
			expectStacks:  []cloudformation.Stack{},
			expectFailure: true,
			failValidate:  true,
		},
		{
			testName:   "Create stack with tags",
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
		{
			testName: "Update stack, adding tags",
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
		{
			testName: "Update stack without tags file, don't remove tags",
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
		{
			testName: "Update stack, remove tags",
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
		{
			testName:           "Create stack with parameters",
			newStackID:         "test-stack/id0",
			parameterInput:     []string{`{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`},
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
		{
			testName:           "Update stack, adding parameters",
			parameterInput:     []string{`{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`},
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
		{
			testName:           "Update stack with subset of parameters",
			parameterInput:     []string{`{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`},
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
		{
			testName:           "Create stack with subset of parameters",
			newStackID:         "test-stack/id0",
			parameterInput:     []string{`{"TestParam1":"TestValue1","TestParam2":"TestValue2"}`},
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
		{
			testName:           "Create stack, missing required parameters",
			newStackID:         "test-stack/id0",
			parameterInput:     []string{`{"TestParam2":"TestValue2"}`},
			requiredParameters: []string{"TestParam1"},
			stacks:             []cloudformation.Stack{},
			expectStacks:       []cloudformation.Stack{},
			expectFailure:      true,
		},
		{
			testName:           "Update stack, missing required parameters",
			parameterInput:     []string{`{"TestParam2":"TestValue2"}`},
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
		{
			testName:    "Create stack, using role",
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
		{
			testName:    "Update stack, using role",
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
		{
			testName:   "Create; JSON Stack Policy",
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
			stackResources: map[string][]resourceValues{
				"test-stack/id0": []resourceValues{
					{
						LogicalResourceId: "ProductionDatabase",
						ResourceType:      "AWS::RDS::DBInstance",
					},
				},
			},
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
		{
			testName: "Update; JSON Stack Policy",
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
			stackResources: map[string][]resourceValues{
				"test-stack/id1": []resourceValues{
					{
						LogicalResourceId: "ProductionDatabase",
						ResourceType:      "AWS::RDS::DBInstance",
					},
				},
			},
			expectStackPolicy: `{"Statement":[{"Action":"Update:*","Effect":"Allow","NotResource":"LogicalResourceId/ProductionDatabase","Principal":"*"}]}`,
		},
		{
			testName:         "Update; Existing JSON Stack Policy without update",
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
		{
			testName:              "Create new stack with termination protection",
			terminationProtection: true,
			newStackID:            "test-stack/id0",
			stacks:                []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id0"),
					StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
		},
		{
			testName:              "Update stack and turn on termination protection",
			terminationProtection: true,
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
					EnableTerminationProtection: aws.Bool(false),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusUpdateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
		},
		{
			testName:              "Update stack and turn on termination protection when no updates to perform",
			noUpdates:             true,
			expectOutput:          DeployOut{Message: "No updates are to be performed."},
			terminationProtection: true,
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
					EnableTerminationProtection: aws.Bool(false),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
		},
		{
			testName: "Update stack and leave termination protection alone",
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusUpdateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
		},
		{
			testName:              "Update stack and leave termination protection alone (v. 2)",
			terminationProtection: false,
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusDeleteComplete),
				},
				{
					StackName:                   aws.String("test-stack"),
					StackId:                     aws.String("test-stack/id1"),
					StackStatus:                 aws.String(cloudformation.StackStatusUpdateComplete),
					EnableTerminationProtection: aws.Bool(true),
				},
			},
		},
		{
			testName:   "Create stack with multiple parameter files",
			newStackID: "test-stack/id0",
			parameterInput: []string{
				`{"One":"Foo","Two":"Foo"}`,
				`{"Two":"Bar"}`,
			},
			requiredParameters: []string{"One", "Two"},
			stacks:             []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("One"), ParameterValue: aws.String("Foo")},
						{ParameterKey: aws.String("Two"), ParameterValue: aws.String("Bar")},
					},
				},
			},
		},
		{
			testName:   "Update stack with multiple parameter files",
			newStackID: "test-stack/id0",
			parameterInput: []string{
				`{"One":"Foo","Two":"Foo"}`,
				`{"Two":"Bar"}`,
			},
			requiredParameters: []string{"One", "Two"},
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("One"), ParameterValue: aws.String("Foo")},
						{ParameterKey: aws.String("Two"), ParameterValue: aws.String("Bar")},
					},
				},
			},
		},
		{
			testName:           "Create stack with parameter overrides",
			newStackID:         "test-stack/id0",
			parameterInput:     []string{`{"One":"Foo","Two":"Foo"}`},
			parameterOverrides: map[string]string{"Two": "Bar"},
			requiredParameters: []string{"One", "Two"},
			stacks:             []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("One"), ParameterValue: aws.String("Foo")},
						{ParameterKey: aws.String("Two"), ParameterValue: aws.String("Bar")},
					},
				},
			},
		},
		{
			testName:           "Update stack with parameter overrides",
			newStackID:         "test-stack/id0",
			parameterInput:     []string{`{"One":"Foo","Two":"Foo"}`},
			parameterOverrides: map[string]string{"Two": "Bar"},
			requiredParameters: []string{"One", "Two"},
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("One"), ParameterValue: aws.String("Foo")},
						{ParameterKey: aws.String("Two"), ParameterValue: aws.String("Bar")},
					},
				},
			},
		},
		{
			testName:           "Create stack with only parameter overrides",
			newStackID:         "test-stack/id0",
			parameterOverrides: map[string]string{"Two": "Bar"},
			requiredParameters: []string{"Two"},
			stacks:             []cloudformation.Stack{},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("Two"), ParameterValue: aws.String("Bar")},
					},
				},
			},
		},
		{
			testName:           "Update stack with parameter overrides",
			newStackID:         "test-stack/id0",
			parameterOverrides: map[string]string{"Two": "Bar"},
			requiredParameters: []string{"Two"},
			stacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusCreateComplete),
				},
			},
			expectStacks: []cloudformation.Stack{
				{
					StackName:   aws.String("test-stack"),
					StackId:     aws.String("test-stack/id0"),
					StackStatus: aws.String(cloudformation.StackStatusUpdateComplete),
					Parameters: []*cloudformation.Parameter{
						{ParameterKey: aws.String("Two"), ParameterValue: aws.String("Bar")},
					},
				},
			},
		},
	}

	oldCFNClient := cfnClient
	defer func() { cfnClient = oldCFNClient }()
	oldSTSClient := stsClient
	defer func() { stsClient = oldSTSClient }()
	for i, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			theseStacks := cases[i].stacks
			theseStackPolicies := c.stackPolicies
			if theseStackPolicies == nil {
				theseStackPolicies = map[string]string{}
			}

			stackResourcesOutput := make(map[string]cloudformation.DescribeStackResourcesOutput)

			for stackName, stackResources := range c.stackResources {

				var cfStackResources []*cloudformation.StackResource
				for _, stackResource := range stackResources {
					cfStackResource := cloudformation.StackResource{
						LogicalResourceId: &stackResource.LogicalResourceId,
						ResourceType:      &stackResource.ResourceType,
					}
					cfStackResources = append(cfStackResources, &cfStackResource)
				}

				stackResourcesOutput[stackName] = cloudformation.DescribeStackResourcesOutput{
					StackResources: cfStackResources,
				}

			}

			cfnClient = mockCfn{
				capabilityIam:        c.capabilityIam,
				failCreate:           c.failCreate,
				failDescribe:         c.failDescribe,
				failValidate:         c.failValidate,
				newStackID:           c.newStackID,
				noUpdates:            c.noUpdates,
				requiredParameters:   c.requiredParameters,
				stacks:               &theseStacks,
				stackResourcesOutput: stackResourcesOutput,
				stackPolicies:        &theseStackPolicies,
			}
			stsClient = mockSTS{accountID: c.accountID}

			thisStack := Stack{
				ParameterBodies:       c.parameterInput,
				ParameterOverrides:    c.parameterOverrides,
				StackName:             "test-stack",
				TagsBody:              c.tagInput,
				TemplateBody:          `{"Resources":{"SNS":{"Type":"AWS::SNS::Topic"},"ProductionDatabase":{"Type":"AWS::RDS::DBInstance"}}}`,
				CfnRoleName:           c.cfnRoleName,
				StackPolicyBody:       c.stackPolicyInput,
				TerminationProtection: c.terminationProtection,
			}

			output, postActions, err := thisStack.Deploy()
			for _, action := range postActions {
				action()
			}

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
				g := theseStackPolicies[thisStack.StackID]
				assert.JSONEq(t, e, g)
			}
		})
	}
}
