package cmd

import "github.com/spf13/cobra"
import "github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
import "fmt"
import "encoding/json"

var PrintManifest = &cobra.Command{
	Use:   "manifest",
	Short: "Show manifest",
	Run: func(cmd *cobra.Command, args []string) {
		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())

		manifest, _ := lib.GetManifest(registry, args)
		buffer, _ := json.MarshalIndent(manifest, "", " ")
		fmt.Println(string(buffer))
	},
}
