package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	forge "github.com/nathandines/forge/forgelib"

	"github.com/spf13/cobra"
)

var stack = forge.Stack{}
var stackInProgressRegexp = regexp.MustCompile("^.*_IN_PROGRESS$")

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge is a CD friendly CloudFormation deployment tool",
	Long: `
Forge is a simple tool which makes deploying CloudFormation stacks a bit more
friendly for continuous delivery environments.
`,
	Version: "v0.1.0-alpha2",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&stack.StackName,
		"stack-name",
		"n",
		"",
		"Name of the stack to manage",
	)
	// rootCmd.PersistentFlags().StringVarP(
	// 	&stack.RoleName,
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

func printStackEvents(s *forge.Stack, after *time.Time) {
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
