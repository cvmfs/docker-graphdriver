package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(garbageCollectionCmd)
}

var garbageCollectionCmd = &cobra.Command{
	Use:     "garbage-collection",
	Short:   "Removes layers that are not necessary anymore",
	Aliases: []string{"gc"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start")
		lib.RemoveUselessLayers()
		os.Exit(0)
	},
}
