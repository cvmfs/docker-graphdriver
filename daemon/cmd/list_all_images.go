package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	listAllImagesCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(listAllImagesCmd)
}

var listAllImagesCmd = &cobra.Command{
	Use:   "list-images",
	Short: "Show all the images actually in the database",
	Run: func(cmd *cobra.Command, args []string) {
		imgs, err := lib.GetAllImagesInDatabase()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for i, img := range imgs {
			img.PrintImage(machineFriendly, i == 0)
		}
		os.Exit(0)
	},
}
