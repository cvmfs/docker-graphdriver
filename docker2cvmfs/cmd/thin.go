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
			fmt.Println("  [docker image] [cvmfs repository]")
			return
		}
		repo := args[1]

		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())

		manifest, _ := lib.GetManifest(registry, args[:1])
		thin := lib.MakeThinImage(manifest, repo)
		j, _ := json.MarshalIndent(thin, "", "  ")
		fmt.Println(string(j))
	},
}
