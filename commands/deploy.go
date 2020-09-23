package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
)

var tagsFile string
var templateFile string
var templateUrl string
var parameterFiles []string
var parameterOverrides []string
var stackPolicyFile string

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		// Read template-file or use template-url

		if templateFile == "" && templateUrl == "" {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("\nArgument 'template-file' or 'template-url' is required\n")
			os.Exit(1)
		}

		if templateFile != "" && templateUrl != "" {
			if err := cmd.Usage(); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("\nEither argument 'template-file' or 'template-url' is required; both cannot be defined\n")
			os.Exit(1)
		}

		if templateFile != "" {
			templateBody, err := ioutil.ReadFile(templateFile)
			if err != nil {
				log.Fatal(err)
			}
			stack.TemplateBody = string(templateBody)
		}
		if templateUrl != "" {
			stack.TemplateUrl = templateUrl
		}

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

		// Parse parameter overrides
		var err error
		stack.ParameterOverrides, err = parseParameterOverrideArgs(parameterOverrides)
		if err != nil {
			log.Fatal(err)
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
			if err := assumeRole(); err != nil {
				log.Fatal("Failed to assume role: ", err)
			}
		}

		output, postActions, err := stack.Deploy()
		if err != nil {
			log.Fatal(err, output)
		}
		for _, postAction := range postActions {
			defer postAction()
		}

		after := stack.LastUpdatedTime

		if after == nil {
			epoch := time.Unix(0, 0)
			after = &epoch
		}

		if t := "No updates are to be performed."; output.Message == t {
			fmt.Println(t)
			return
		}

		for {
		refresh_stack_status:
			if err := stack.GetStackInfo(); err != nil {
				if assumeRoleArn == "" {
					log.Fatal(err)
				}
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
		"Path to the CloudFormation template to be deployed.  Must not specify template-url.",
	)
	deployCmd.MarkFlagFilename("template-file")

	deployCmd.PersistentFlags().StringVar(
		&templateUrl,
		"template-url",
		"",
		"S3 url to the CloudFormation template to be deployed.  Must not specify template-file.",
	)

	deployCmd.PersistentFlags().StringSliceVarP(
		&parameterFiles,
		"parameters-file",
		"p",
		[]string{},
		"Path to the file which contains the parameters for this stack. Can be defined multiple\n"+
			"times to merge files, later ones overriding earlier ones.",
	)
	deployCmd.MarkFlagFilename("parameters-file")

	deployCmd.PersistentFlags().StringSliceVarP(
		&parameterOverrides,
		"parameter-override",
		"o",
		[]string{},
		"Overrides for parameters (format \"<key>=<value>\"). Can be defined multiple times for\n"+
			"multiple overrides.",
	)

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
		"Path to the file which contains the stack policy for this stack and all nested stack",
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
