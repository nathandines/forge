package forgelib

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type mockCfn struct {
	capabilityIam        bool
	failCreate           bool
	failDescribe         bool
	failValidate         bool
	newStackID           string
	noUpdates            bool
	requiredParameters   []string
	stackEventsOutput    cloudformation.DescribeStackEventsOutput
	stackResourcesOutput map[string]cloudformation.DescribeStackResourcesOutput
	stackPolicies        *map[string]string
	stacks               *[]cloudformation.Stack
	cloudformationiface.CloudFormationAPI
}

func checkIamCapability(inputCapabilities []*string) (err error) {
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

func (m mockCfn) ValidateTemplate(*cloudformation.ValidateTemplateInput) (*cloudformation.ValidateTemplateOutput, error) {
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
	for _, r := range m.requiredParameters {
		thisParameter := cloudformation.TemplateParameter{ParameterKey: aws.String(r)}
		output.Parameters = append(output.Parameters, &thisParameter)
	}
	return &output, nil
}

func (m mockCfn) CreateStack(input *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error) {
	output := cloudformation.CreateStackOutput{}

	if m.failCreate {
		return &output, awserr.New(
			cloudformation.ErrCodeInvalidOperationException,
			"Simulated Failure",
			nil,
		)
	}

	// Fail if IAM capabilities are required, but not supplied
	if m.capabilityIam {
		if err := checkIamCapability(input.Capabilities); err != nil {
			return &output, err
		}
	}

	// Check that all required parameters are supplied
REQUIRED_PARAMETERS:
	for _, r := range m.requiredParameters {
		for _, s := range input.Parameters {
			if r == *s.ParameterKey {
				continue REQUIRED_PARAMETERS
			}
		}
		return &output, awserr.New(
			"ValidationError",
			fmt.Sprintf("Parameters: [%s] must have values", r),
			nil,
		)
	}

	// Check for existing stack, fail if found
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
		StackName:                   input.StackName,
		StackId:                     aws.String(m.newStackID),
		StackStatus:                 aws.String(cloudformation.StackStatusCreateComplete),
		Tags:                        input.Tags,
		Parameters:                  input.Parameters,
		RoleARN:                     input.RoleARN,
		EnableTerminationProtection: input.EnableTerminationProtection,
	}
	*m.stacks = append(*m.stacks, thisStack)

	if input.StackPolicyBody != nil {
		(*m.stackPolicies)[m.newStackID] = *input.StackPolicyBody
	}

	output.StackId = &m.newStackID
	return &output, nil
}

func (m mockCfn) UpdateStack(input *cloudformation.UpdateStackInput) (*cloudformation.UpdateStackOutput, error) {
	output := cloudformation.UpdateStackOutput{}
	if m.capabilityIam {
		if err := checkIamCapability(input.Capabilities); err != nil {
			return &output, err
		}
	}

	// Throw error if no updates are to be performed for stack
	if m.noUpdates {
		return &output, awserr.New(
			"ValidationError",
			"No updates are to be performed.",
			nil,
		)
	}

	// Check that all required parameters are supplied
REQUIRED_PARAMETERS:
	for _, r := range m.requiredParameters {
		for _, s := range input.Parameters {
			if r == *s.ParameterKey {
				continue REQUIRED_PARAMETERS
			}
		}
		return &output, awserr.New(
			"ValidationError",
			fmt.Sprintf("Parameters: [%s] must have values", r),
			nil,
		)
	}

	// For each existing stack, match against the stack ID first, then the stack
	// name. If found and the stack is in a good state, set values against it
	// and return
	for i := 0; i < len(*m.stacks); i++ {
		if s := *input.StackName; s == *(*m.stacks)[i].StackId ||
			s == *(*m.stacks)[i].StackName {
			switch *(*m.stacks)[i].StackStatus {
			case cloudformation.StackStatusCreateComplete,
				cloudformation.StackStatusUpdateComplete,
				cloudformation.StackStatusUpdateRollbackComplete:

				*(*m.stacks)[i].StackStatus = cloudformation.StackStatusUpdateComplete
				(*m.stacks)[i].RoleARN = input.RoleARN
				(*m.stacks)[i].Tags = input.Tags
				(*m.stacks)[i].Parameters = input.Parameters

				if input.StackPolicyBody != nil {
					(*m.stackPolicies)[*(*m.stacks)[i].StackId] = *input.StackPolicyBody
				}

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

func (m mockCfn) DescribeStacks(input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
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

func (m mockCfn) DeleteStack(input *cloudformation.DeleteStackInput) (output *cloudformation.DeleteStackOutput, err error) {
	for i := 0; i < len(*m.stacks); i++ {
		if *(*m.stacks)[i].StackId == *input.StackName &&
			*(*m.stacks)[i].StackStatus != cloudformation.StackStatusDeleteComplete {
			*(*m.stacks)[i].StackStatus = cloudformation.StackStatusDeleteComplete
			(*m.stacks)[i].RoleARN = input.RoleARN
			return
		}
	}
	return
}

func (m mockCfn) DescribeStackEventsPages(input *cloudformation.DescribeStackEventsInput, function func(*cloudformation.DescribeStackEventsOutput, bool) bool) error {
	// Paginate events to test that the destination functions concatenate the
	// entries correctly
	for i := 0; i < len(m.stackEventsOutput.StackEvents); i++ {
		thisOutput := &cloudformation.DescribeStackEventsOutput{
			StackEvents: []*cloudformation.StackEvent{
				m.stackEventsOutput.StackEvents[i],
			},
		}
		if nextPage := function(thisOutput, true); !nextPage {
			return nil
		}
	}
	return nil
}

func (m mockCfn) DescribeStackResources(input *cloudformation.DescribeStackResourcesInput) (*cloudformation.DescribeStackResourcesOutput, error) {

	output := m.stackResourcesOutput[*input.StackName]
	return &output, nil
}

func (m mockCfn) SetStackPolicy(input *cloudformation.SetStackPolicyInput) (*cloudformation.SetStackPolicyOutput, error) {

	output := cloudformation.SetStackPolicyOutput{}
	if input.StackPolicyBody != nil {
		(*m.stackPolicies)[*input.StackName] = *input.StackPolicyBody
	}
	return &output, nil
}

func (m mockCfn) UpdateTerminationProtection(input *cloudformation.UpdateTerminationProtectionInput) (*cloudformation.UpdateTerminationProtectionOutput, error) {
	output := cloudformation.UpdateTerminationProtectionOutput{}
	// For each existing stack, match against the stack ID first, then the stack
	// name. If found and the stack is in a good state, set values against it
	// and return
STACK_LOOP:
	for i := 0; i < len(*m.stacks); i++ {
		if s := *input.StackName; s == *(*m.stacks)[i].StackId ||
			s == *(*m.stacks)[i].StackName {
			switch *(*m.stacks)[i].StackStatus {
			case cloudformation.StackStatusDeleteComplete,
				cloudformation.StackStatusDeleteInProgress:
				break STACK_LOOP
			default:
				(*m.stacks)[i].EnableTerminationProtection = input.EnableTerminationProtection
				output.StackId = (*m.stacks)[i].StackId
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
