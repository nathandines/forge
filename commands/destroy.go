package commands

import (
	"encoding/json"
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
			bunch, err := stackResource.ListEvents(after)
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

			switch *stackResource.StackInfo.StackStatus {
			case cloudformation.StackStatusDeleteComplete:
				os.Exit(0)
			case cloudformation.StackStatusDeleteFailed:
				os.Exit(1)
			}

			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
