package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	forge "github.com/nathandines/forge/forgelib"
	"github.com/spf13/cobra"
)

var tagsFile string
var templateFile string
var parameterFiles []string
var stackPolicyFile string

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		// Read template-file
		if templateFile == "" {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("\nArgument 'template-file' is required\n")
			os.Exit(1)
		}
		templateBody, err := ioutil.ReadFile(templateFile)
		if err != nil {
			log.Fatal(err)
		}
		stack.TemplateBody = string(templateBody)

		// Read tags-file
		if tagsFile != "" {
			tagsBody, err := ioutil.ReadFile(tagsFile)
			if err != nil {
				log.Fatal(err)
			}
			stack.TagsBody = string(tagsBody)
		}

		// Read parameters-file
		for _, p := range parameterFiles {
			parametersBody, err := ioutil.ReadFile(p)
			if err != nil {
				log.Fatal(err)
			}
			stack.ParameterBodies = append(stack.ParameterBodies, string(parametersBody))
		}

		// Read stack-policy-file
		if stackPolicyFile != "" {
			stackPolicyBody, err := ioutil.ReadFile(stackPolicyFile)
			if err != nil {
				log.Fatal(err)
			}
			stack.StackPolicyBody = string(stackPolicyBody)
		}

		if assumeRoleArn != "" {
			if err := forge.AssumeRole(assumeRoleArn); err != nil {
				log.Fatal(err)
			}
		}

		// Populate Stack ID
		// Deliberately ignore errors here, as the stack might not exist yet
		stack.GetStackInfo()

		after, err := stack.GetLastEventTime()
		if err != nil {
			// default to epoch as the time to look for events from
			epoch := time.Unix(0, 0)
			after = &epoch
		}

		output, err := stack.Deploy()
		if err != nil {
			log.Fatal(err)
		}

		if t := "No updates are to be performed."; output.Message == t {
			fmt.Println(t)
			return
		}

		for {
		refresh_stack_status:
			if err := stack.GetStackInfo(); err != nil {
				if err2 := rotateRoleCredentials(err); err2 != nil {
					log.Fatal(err)
				}
				goto refresh_stack_status
			}

			printStackEvents(&stack, after)

			status := *stack.StackInfo.StackStatus
			switch {
			case stackInProgressRegexp.MatchString(status):
			case status == cloudformation.StackStatusCreateComplete:
				return
			case status == cloudformation.StackStatusUpdateComplete:
				return
			default:
				fmt.Print("\n")
				log.Fatal(fmt.Errorf("Stack deploy failed! Stack Status: %s", status))
			}

			time.Sleep(time.Duration(eventPollingPeriod) * time.Second)
		}
	},
}

func init() {
	deployCmd.PersistentFlags().StringVarP(
		&templateFile,
		"template-file",
		"t",
		"",
		"Path to the CloudFormation template to be deployed",
	)
	deployCmd.MarkFlagFilename("template-file")

	deployCmd.PersistentFlags().StringSliceVarP(
		&parameterFiles,
		"parameters-file",
		"p",
		[]string{},
		"Path to the file which contains the parameters for this stack (can be defined multiple times for overrides)",
	)
	deployCmd.MarkFlagFilename("parameters-file")

	deployCmd.PersistentFlags().StringVar(
		&tagsFile,
		"tags-file",
		"",
		"Path to the file which contains the tags for this stack",
	)
	deployCmd.MarkFlagFilename("tags-file")

	deployCmd.PersistentFlags().StringVar(
		&stackPolicyFile,
		"stack-policy-file",
		"",
		"Path to the file which contains the stack policy for this stack",
	)
	deployCmd.MarkFlagFilename("stack-policy-file")

	deployCmd.PersistentFlags().BoolVar(
		&stack.TerminationProtection,
		"termination-protection",
		false,
		"Set termination protection for this stack",
	)
	deployCmd.MarkFlagFilename("stack-policy-file")

	rootCmd.AddCommand(deployCmd)
}
