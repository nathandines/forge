package commands

import "github.com/spf13/cobra"

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a CloudFormation Stack",
	Run: func(cmd *cobra.Command, args []string) {
		stackResource.Destroy()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
