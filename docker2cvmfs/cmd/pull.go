package cmd

import "github.com/spf13/cobra"
import "github.com/cvmfs/docker2cvmfs/docker2cvmfs/lib"

var PullLayers = &cobra.Command {
	Use: "pull layers",
	Short: "pull the layers",
	Run: func(cmd *cobra.Command, args []string) {
		lib.PullLayers(args)
	},
}
