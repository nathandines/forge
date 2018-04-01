package commands

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		// Populate Stack ID
		// Deliberately ignore errors here
		stackResource.GetStackInfo()

		after, err := stackResource.GetLastEventTime()
		if err != nil {
			// default to epoch as the time to look for events from
			epoch := time.Unix(0, 0)
			after = &epoch
		}

		if err := stackResource.Deploy(); err != nil {
			log.Fatal(err)
		}

		for {
			// Refresh Stack State
			if err := stackResource.GetStackInfo(); err != nil {
				log.Fatal(err)
			}

			printStackEvents(&stackResource, after)

			status := *stackResource.StackInfo.StackStatus
			switch {
			case stackInProgressRegexp.MatchString(status):
			case status == cloudformation.StackStatusCreateComplete:
				os.Exit(0)
			case status == cloudformation.StackStatusUpdateComplete:
				os.Exit(0)
			default:
				os.Exit(1)
			}

			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	deployCmd.PersistentFlags().StringVar(
		&stackResource.StackPolicyFile,
		"stack-policy-file",
		"",
		`Path to the stack policy which should be applied to this CloudFormation
stack`,
	)
	deployCmd.MarkFlagFilename("stack-policy-file")

	deployCmd.PersistentFlags().StringVarP(
		&stackResource.TemplateFile,
		"template-file",
		"t",
		"",
		"Path to the CloudFormation template to be deployed",
	)
	deployCmd.MarkFlagFilename("template-file")

	deployCmd.PersistentFlags().StringVarP(
		&stackResource.ParametersFile,
		"parameters-file",
		"p",
		"",
		"Path to the file which contains the parameters for this stack",
	)
	deployCmd.MarkFlagFilename("parameters-file")
	rootCmd.AddCommand(deployCmd)
}
