package cmd

import "github.com/spf13/cobra"
import "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"

var PullLayers = &cobra.Command{
	Use:   "pull layers",
	Short: "pull the layers",
	Run: func(cmd *cobra.Command, args []string) {
		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())
		inputReference := args[0]
		lib.PullLayers(registry, inputReference, "docker.cern.ch", "layers/")
	},
}
