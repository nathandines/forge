package commands

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		// Populate Stack ID
		if err := stackResource.GetStackInfo(); err != nil {
			log.Fatal(err)
		}

		after, err := stackResource.GetLastEventTime()
		if err != nil {
			log.Fatal(err)
		}

		if err := stackResource.Destroy(); err != nil {
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
			case status == cloudformation.StackStatusDeleteComplete:
				return
			default:
				fmt.Print("\n")
				log.Fatal(fmt.Errorf("Stack destroy failed! Stack Status: %s", status))
			}

			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
