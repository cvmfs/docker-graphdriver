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

var repository string
var input_docker_reference string
var output_docker_reference string
var subdirectory string

func init() {
	RootCmd.PersistentFlags().String("registry", "https://registry-1.docker.io/v2", "Docker registry url")
	RootCmd.AddCommand(PullLayers)
	RootCmd.AddCommand(PrintManifest)
	RootCmd.AddCommand(PrintConfig)
	RootCmd.AddCommand(CreateThinImage)
	RootCmd.AddCommand(MakeThin)

	MakeThin.Flags().StringVarP(&input_docker_reference, "input-reference", "i", "", "Input reference [REPOSITORY:TAG] to make thin.")
	MakeThin.Flags().StringVarP(&output_docker_reference, "output-reference", "o", "", "Repository and tag to use for the output reference")
	MakeThin.Flags().StringVarP(&repository, "repository", "r", "", "Repository where to store the docker layers.")
	MakeThin.Flags().StringVarP(&subdirectory, "subdirectory", "s", "layers/", "In which subdirectory store the docker layers inside the repository.")
	MakeThin.MarkFlagRequired("input-reference")
	MakeThin.MarkFlagRequired("output-reference")
	MakeThin.MarkFlagRequired("repository")
}
