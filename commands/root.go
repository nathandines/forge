package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	forge "github.com/nathandines/forge/v2/forgelib"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/spf13/cobra"
)

var stack = forge.Stack{}
var stackInProgressRegexp = regexp.MustCompile("^.*_IN_PROGRESS$")

var assumeRoleArn string
var assumeRoleMFASerial string
var assumeRoleWithMFA bool
var eventPollingPeriod int

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge is a CD friendly CloudFormation deployment tool",
	Long: `
Forge is a simple tool which makes deploying CloudFormation stacks a bit more
friendly for continuous delivery environments.

GitHub: https://github.com/nathandines/forge
`,
	Version: "v2.3.0",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&stack.StackName,
		"stack-name",
		"n",
		"",
		"Name of the stack to manage",
	)
	rootCmd.PersistentFlags().StringVar(
		&stack.CfnRoleName,
		"cfn-role-name",
		"",
		"Name of IAM role in the destination account for the CloudFormation service to assume",
	)
	rootCmd.PersistentFlags().StringVar(
		&assumeRoleArn,
		"assume-role-arn",
		"",
		"Name of IAM role to assume BEFORE making requests to CloudFormation",
	)
	rootCmd.PersistentFlags().BoolVar(
		&assumeRoleWithMFA,
		"assume-role-with-mfa",
		false,
		"Flag to specify that MFA is required to assume the role",
	)
	rootCmd.PersistentFlags().StringVar(
		&assumeRoleMFASerial,
		"assume-role-mfa-serial",
		"",
		"Specify the MFA serial if it cannot be automatically detected",
	)
	rootCmd.PersistentFlags().IntVar(
		&eventPollingPeriod,
		"event-polling-period",
		10,
		"Polling period in seconds for monitoring CloudFormation stack events",
	)
}

// Execute does what it says on the box
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printStackEvents(s *forge.Stack, after *time.Time) {
list_events:
	bunch, err := s.ListEvents(after)
	if err != nil {
		if err2 := rotateRoleCredentials(err); err2 != nil {
			log.Fatal(err)
		}
		goto list_events
	}
	for _, e := range bunch {
		// IDs renamed for JSON output to match the API response data
		stackEvent := struct {
			LogicalResourceID    *string   `json:"LogicalResourceId"`
			PhysicalResourceID   *string   `json:"PhysicalResourceId,omitempty"`
			ResourceStatus       *string   `json:""`
			ResourceStatusReason *string   `json:",omitempty"`
			ResourceType         *string   `json:""`
			Timestamp            time.Time `json:""`
		}{
			(*e).LogicalResourceId,
			(*e).PhysicalResourceId,
			(*e).ResourceStatus,
			(*e).ResourceStatusReason,
			(*e).ResourceType,
			(*(*e).Timestamp).Local(),
		}
		jsonData, err := json.MarshalIndent(stackEvent, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonData))
	}
	if len(bunch) > 0 {
		*after = *bunch[len(bunch)-1].Timestamp
	}
}

func rotateRoleCredentials(err error) error {
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
		case "ExpiredToken":
			forge.UnassumeAllRoles()
			if err2 := assumeRole(); err2 != nil {
				return err2
			}
		default:
			return err
		}
	} else {
		return err
	}
	return nil
}

func assumeRole() error {
	if assumeRoleWithMFA {
		mfaToken, err := stscreds.StdinTokenProvider()
		if err != nil {
			return err
		}
		if err := forge.AssumeRoleWithMFA(assumeRoleArn, mfaToken, assumeRoleMFASerial); err != nil {
			return err
		}
	} else {
		if err := forge.AssumeRole(assumeRoleArn); err != nil {
			return err
		}
	}
	return nil
}
