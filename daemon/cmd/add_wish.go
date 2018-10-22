package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	inputImage, outputImage, cvmfsRepo, userInput, userOutput string
	convert                                                   bool
)

func init() {
	addWishCmd.Flags().StringVarP(&inputImage, "input-image", "i", "", "input image to add to the wish triplet")
	addWishCmd.MarkFlagRequired("input-image")

	addWishCmd.Flags().StringVarP(&outputImage, "output-image", "o", "", "output image to add to the wish triplet")
	addWishCmd.MarkFlagRequired("output-image")

	addWishCmd.Flags().StringVarP(&cvmfsRepo, "repository", "r", "", "cvmfs repository add to the wish triplet")
	addWishCmd.Flags().StringVarP(&userInput, "user-input", "a", "", "username to access the input registry")
	addWishCmd.Flags().StringVarP(&userOutput, "user-output", "b", "", "username to access the output registry")
	addWishCmd.MarkFlagRequired("repository")

	addWishCmd.Flags().BoolVarP(&convert, "convert", "c", false, "start the conversion process immediately after adding the wish")
	addWishCmd.Flags().BoolVarP(&overwriteLayer, "overwrite-layers", "f", false, "overwrite the layer if they are already inside the CVMFS repository")
	addWishCmd.Flags().BoolVarP(&convertSingularity, "convert-singularity", "s", true, "also create a singularity images")

	rootCmd.AddCommand(addWishCmd)
}

var addWishCmd = &cobra.Command{
	Use:   "add-wish",
	Short: "Add a wish triplet",
	Run: func(cmd *cobra.Command, args []string) {
		wish, err := lib.CreateWish(inputImage, outputImage, cvmfsRepo, userInput, userOutput)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to create the wish")
		}
		// at this point we add the real wish to the database
		err = lib.AddWish(wish.InputImage, wish.OutputImage, wish.CvmfsRepo)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to add the wish to the database")
		}

		// if required to convert mmediately the wish we do so
		if convert {
			AliveMessage()
			wish, err := lib.GetWishF(wish.InputImage, wish.OutputImage, wish.CvmfsRepo)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to retrieve the wish just added")
			}
			err = lib.ConvertWish(wish, false, overwriteLayer, convertSingularity)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to convert the newly added wish")
			}
		}
	},
}
