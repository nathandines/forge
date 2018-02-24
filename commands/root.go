package commands

import (
	"fmt"
	"os"

	stack "github.com/nathandines/stack/stacklib"

	"github.com/spf13/cobra"
)

var stackResource = stack.Stack{}

var rootCmd = &cobra.Command{
	Use:   "stack",
	Short: "Stack is a CD friendly CloudFormation deployment tool",
	Long: `
Stack is a simple tool which makes deploying CloudFormation stacks a bit more
friendly for continuous delivery environments.
`,
	Version: "v0.1.0",
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&stackResource.ProjectManifest,
		"project-manifest",
		"project.yml",
		"Path to the project manifest",
	)
	rootCmd.MarkFlagFilename("project-manifest")

	rootCmd.PersistentFlags().StringVarP(
		&stackResource.StackName,
		"stack-name",
		"n",
		"",
		"Name of the stack to manage",
	)
	rootCmd.PersistentFlags().StringVarP(
		&stackResource.RoleName,
		"role-name",
		"r",
		"",
		"Name of IAM role in this account for CloudFormation to assume",
	)
}

// Execute does what it says on the box
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
