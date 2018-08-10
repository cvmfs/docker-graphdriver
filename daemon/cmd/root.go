package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&lib.DatabaseLocation, "database", "d", lib.DefaultDatabaseLocation, "database location")
}

var rootCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Show the several commands available.",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func EntryPoint() {
	rootCmd.Execute()
}
