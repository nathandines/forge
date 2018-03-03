package commands

import (
	"log"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		if err := stackResource.Deploy(); err != nil {
			log.Fatal(err)
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
