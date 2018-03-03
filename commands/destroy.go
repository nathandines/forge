package commands

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		// var events []*cloudformation.StackEvent
		eventStart := time.Now()
		if err := stackResource.Destroy(); err != nil {
			log.Fatal(err)
		}
		for {
			eventEnd := time.Now()
			bunch, err := stackResource.Events(&eventStart, &eventEnd)
			if err != nil {
				log.Fatal(err)
			}
			for _, e := range bunch {
				// events = append(events, e)
				var statusReason string
				if e.ResourceStatusReason != nil {
					statusReason = *e.ResourceStatusReason
				} else {
					statusReason = "<undefined>"
				}
				fmt.Printf(
					"%s: %s; %s; %s; %s; Reason: %s\n",
					e.Timestamp.Local(),
					*e.ResourceStatus,
					*e.ResourceType,
					*e.LogicalResourceId,
					*e.PhysicalResourceId,
					statusReason,
				)
			}
			eventStart = eventEnd

			if len(bunch) > 0 && *bunch[len(bunch)-1].PhysicalResourceId == stackResource.StackID {
				switch *bunch[len(bunch)-1].ResourceStatus {
				case cloudformation.StackStatusDeleteComplete:
					os.Exit(0)
				case cloudformation.StackStatusDeleteFailed:
					os.Exit(1)
				}
			}
			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
