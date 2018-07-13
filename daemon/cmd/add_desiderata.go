package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	inputImage, outputImage, cvmfsRepo string
)

func init() {
	addDesiderataCmd.Flags().StringVarP(&inputImage, "input-image", "i", "", "input image to add to the desiderata triplet")
	addDesiderataCmd.MarkFlagRequired("input-image")

	addDesiderataCmd.Flags().StringVarP(&outputImage, "output-image", "o", "", "output image to add to the desiderata triplet")
	addDesiderataCmd.MarkFlagRequired("output-image")

	addDesiderataCmd.Flags().StringVarP(&cvmfsRepo, "repository", "r", "", "cvmfs repository add to the desiderata triplet")
	addDesiderataCmd.MarkFlagRequired("repository")

	rootCmd.AddCommand(addDesiderataCmd)
}

var addDesiderataCmd = &cobra.Command{
	Use:   "add-desiderata",
	Short: "Add a desiderata triplet",
	Run: func(cmd *cobra.Command, args []string) {
		inputImg, err := lib.ParseImage(inputImage)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		outputImg, err := lib.ParseImage(outputImage)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(inputImg, outputImg)
	},
}
