package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	stack "github.com/nathandines/stack/stacklib"

	"github.com/spf13/cobra"
)

var stackResource = stack.Stack{}
var stackInProgressRegexp = regexp.MustCompile("^.*_IN_PROGRESS$")

var rootCmd = &cobra.Command{
	Use:   "stack",
	Short: "Stack is a CD friendly CloudFormation deployment tool",
	Long: `
Stack is a simple tool which makes deploying CloudFormation stacks a bit more
friendly for continuous delivery environments.
`,
	Version: "v0.1.0-alpha2",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&stackResource.StackName,
		"stack-name",
		"n",
		"",
		"Name of the stack to manage",
	)
	// rootCmd.PersistentFlags().StringVarP(
	// 	&stackResource.RoleName,
	// 	"role-name",
	// 	"r",
	// 	"",
	// 	"Name of IAM role in this account for CloudFormation to assume",
	// )
}

// Execute does what it says on the box
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printStackEvents(s *stack.Stack, after *time.Time) {
	bunch, err := s.ListEvents(after)
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range bunch {
		jsonData, err := json.MarshalIndent(*e, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonData))
	}
	if len(bunch) > 0 {
		*after = *bunch[len(bunch)-1].Timestamp
	}
}
