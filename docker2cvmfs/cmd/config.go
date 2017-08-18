package cmd

import "github.com/spf13/cobra"
import "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"

var PrintConfig = &cobra.Command{
	Use:   "config",
	Short: "Print image config",
	Run: func(cmd *cobra.Command, args []string) {
		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())
		lib.GetConfig(registry, args)
	},
}
