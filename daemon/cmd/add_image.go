package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	addImageCmd.Flags().StringVarP(&username, "username", "u", "", "username to use to log in into the registry.")
	addImageCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(addImageCmd)
}

var addImageCmd = &cobra.Command{
	Use:   "add-image",
	Short: "Add an image to the database, every image added is to be classified as a not thin image",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		img, err := lib.ParseImage(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = img.GetManifest()
		if err != nil {
			lib.LogE(err).Fatal("Impossible to get the manifest of the image, wrong registry or missing credential maybe?")
		}
		err = lib.AddImage(img)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		img.PrintImage(machineFriendly, true)
		os.Exit(0)
	},
}
