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

var templateFile string
var tagsFile string

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
		stackResource.TemplateBody = string(templateBody)

		// Read tags-file
		if tagsFile != "" {
			tagsBody, err := ioutil.ReadFile(tagsFile)
			if err != nil {
				log.Fatal(err)
			}
			stackResource.TagsBody = string(tagsBody)
		}

		// Populate Stack ID
		// Deliberately ignore errors here
		stackResource.GetStackInfo()

		after, err := stackResource.GetLastEventTime()
		if err != nil {
			// default to epoch as the time to look for events from
			epoch := time.Unix(0, 0)
			after = &epoch
		}

		output, err := stackResource.Deploy()
		if err != nil {
			log.Fatal(err)
		}

		if t := "No updates are to be performed."; output.Message == t {
			fmt.Println(t)
			return
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
				return
			case status == cloudformation.StackStatusUpdateComplete:
				return
			default:
				os.Exit(1)
			}

			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	// deployCmd.PersistentFlags().StringVar(
	// 	&stackResource.StackPolicyFile,
	// 	"stack-policy-file",
	// 	"",
	// 	"Path to the stack policy which should be applied to this CloudFormation stack",
	// )
	// deployCmd.MarkFlagFilename("stack-policy-file")

	deployCmd.PersistentFlags().StringVarP(
		&templateFile,
		"template-file",
		"t",
		"",
		"Path to the CloudFormation template to be deployed",
	)
	deployCmd.MarkFlagFilename("template-file")

	// deployCmd.PersistentFlags().StringVarP(
	// 	&stackResource.ParametersFile,
	// 	"parameters-file",
	// 	"p",
	// 	"",
	// 	"Path to the file which contains the parameters for this stack",
	// )
	// deployCmd.MarkFlagFilename("parameters-file")

	deployCmd.PersistentFlags().StringVar(
		&tagsFile,
		"tags-file",
		"",
		"Path to the file which contains the tags for this stack",
	)
	deployCmd.MarkFlagFilename("tags-file")

	rootCmd.AddCommand(deployCmd)
}
