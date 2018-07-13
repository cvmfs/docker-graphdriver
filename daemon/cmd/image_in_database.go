package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(imageInDatabaseCmd)
}

var imageInDatabaseCmd = &cobra.Command{
	Use:   "image-in-database",
	Short: "Check that the provide image is already in the database.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		img, err := lib.ParseImage(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if lib.IsImageInDatabase(img) {
			fmt.Println(1)
		} else {
			fmt.Println(0)
		}
		os.Exit(0)
	},
}
