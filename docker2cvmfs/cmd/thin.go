package cmd

import "fmt"
import "encoding/json"

import "github.com/spf13/cobra"
import "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"

var CreateThinImage = &cobra.Command{
	Use:   "thin",
	Short: "Make thin image.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("Error: invalid arguments.")
			fmt.Println("  [docker image] [cvmfs repository + path prefix]")
			fmt.Println("  Example: library/ubuntu:latest images.cern.ch/layers")
			return
		}
		repoLocation := args[1]

		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())

		manifest, _ := lib.GetManifest(registry, args[:1])
		origin := args[0] + "@" + registry
		thin := lib.MakeThinImage(manifest, repoLocation, origin)
		j, _ := json.MarshalIndent(thin, "", "  ")
		fmt.Println(string(j))
	},
}
