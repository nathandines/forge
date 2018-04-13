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
var parametersFile string

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

		// Read tags-file
		if parametersFile != "" {
			parametersBody, err := ioutil.ReadFile(parametersFile)
			if err != nil {
				log.Fatal(err)
			}
			stack.ParametersBody = string(parametersBody)
		}

		// Populate Stack ID
		// Deliberately ignore errors here
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
			// Refresh Stack State
			if err := stack.GetStackInfo(); err != nil {
				log.Fatal(err)
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

			time.Sleep(5 * time.Second)
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

	deployCmd.PersistentFlags().StringVarP(
		&parametersFile,
		"parameters-file",
		"p",
		"",
		"Path to the file which contains the parameters for this stack",
	)
	deployCmd.MarkFlagFilename("parameters-file")

	deployCmd.PersistentFlags().StringVar(
		&tagsFile,
		"tags-file",
		"",
		"Path to the file which contains the tags for this stack",
	)
	deployCmd.MarkFlagFilename("tags-file")

	rootCmd.AddCommand(deployCmd)
}
