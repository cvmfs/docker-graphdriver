package cmd

import "fmt"
import "github.com/spf13/cobra"

var RootCmd = &cobra.Command {
	Use: "docker2cvmfs",
	Short: "Human friendly interaction with the docker hub...",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello world!")
	},
}

func init() {
	RootCmd.AddCommand(PullLayers)
	RootCmd.AddCommand(PrintManifest)
	RootCmd.AddCommand(PrintConfig)
}
