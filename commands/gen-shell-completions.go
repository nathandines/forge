package commands

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var genBashCompletion = &cobra.Command{
	Use:   "gen-bash-completion",
	Short: "Generate bash completion file output",
	Run: func(cmd *cobra.Command, args []string) {
		if err := rootCmd.GenBashCompletion(os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(genBashCompletion)
}

var genZshCompletion = &cobra.Command{
	Use:   "gen-zsh-completion",
	Short: "Generate zsh completion file output",
	Run: func(cmd *cobra.Command, args []string) {
		if err := rootCmd.GenZshCompletion(os.Stdout); err != nil {
			log.Fatal(err)
		}
	},
	Hidden: true,
}

func init() {
	rootCmd.AddCommand(genZshCompletion)
}
