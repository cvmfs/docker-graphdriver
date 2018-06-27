package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
	"github.com/spf13/cobra"
)

var PrintManifest = &cobra.Command{
	Use:   "manifest",
	Short: "Show manifest",
	Run: func(cmd *cobra.Command, args []string) {
		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())

		manifest, _ := lib.GetManifest(registry, args[0])
		buffer, _ := json.MarshalIndent(manifest, "", " ")
		fmt.Println(string(buffer))
	},
}
