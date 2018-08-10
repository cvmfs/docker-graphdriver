package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	convertAgain, overwriteLayer bool
)

func init() {
	convertCmd.Flags().BoolVarP(&overwriteLayer, "overwrite-layers", "f", false, "overwrite the layer if they are already inside the CVMFS repository")
	convertCmd.Flags().BoolVarP(&convertAgain, "convert-again", "g", false, "convert again images that are already successfull converted")
	rootCmd.AddCommand(convertCmd)
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert the wishes",
	Run: func(cmd *cobra.Command, args []string) {
		wish, err := lib.GetAllWishes()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, wish := range wish {
			err = lib.ConvertWish(wish, convertAgain, overwriteLayer)
			if err != nil {
				fmt.Println(err)
			}
		}
		os.Exit(0)
	},
}