package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Show the several commands available.",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func EntryPoint() {
	rootCmd.Execute()
}
