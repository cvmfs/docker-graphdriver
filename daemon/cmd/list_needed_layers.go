package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	listNeededLayers.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(listNeededLayers)
}

var listNeededLayers = &cobra.Command{
	Use:     "list-needed-layers",
	Short:   "Show all the layers that are needed",
	Aliases: []string{"list-needed-layer", "ls-layer", "ls-layers", "layer-ls", "layers-ls", "list-layers", "list-layer"},
	Run: func(cmd *cobra.Command, args []string) {
		layers, err := lib.GetAllNeededLayers()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, layer := range layers {
			fmt.Println(layer)
		}
		os.Exit(0)
	},
}
