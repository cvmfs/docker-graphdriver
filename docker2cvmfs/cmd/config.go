package cmd

import "github.com/spf13/cobra"
import "github.com/cvmfs/docker2cvmfs/docker2cvmfs/lib"

var PrintConfig = &cobra.Command {
	Use: "config",
	Short: "Print image config",
	Run: func(cmd *cobra.Command, args []string) {
		lib.GetConfig(args)
	},
}
