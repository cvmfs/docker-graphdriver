package cmd

import "fmt"
import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "docker2cvmfs",
	Short: "Human friendly interaction with the docker hub...",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use `docker2cvmfs help` for a list of supported commands.")
		fmt.Println("Registry: ")
		fmt.Println("    " + cmd.PersistentFlags().Lookup("registry").Value.String())
	},
}

func init() {
	RootCmd.PersistentFlags().String("registry", "https://registry-1.docker.io/v2", "Docker registry url")

	RootCmd.AddCommand(PullLayers)
	RootCmd.AddCommand(PrintManifest)
	RootCmd.AddCommand(PrintConfig)
	RootCmd.AddCommand(CreateThinImage)

}
