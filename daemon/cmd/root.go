package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var rootCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Show the several commands available.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func EntryPoint() {
	rootCmd.Execute()
}

func AliveMessage() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for _ = range ticker.C {
			lib.Log().Info("Process alive")
		}
	}()
}
