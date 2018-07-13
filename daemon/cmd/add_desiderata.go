package cmd

import (
	"github.com/spf13/cobra"
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
	Run:   func(cmd *cobra.Command, args []string) {},
}
