package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	inputImage, outputImage, cvmfsRepo, userInput, userOutput string
	convert                                                   bool
)

func init() {
	addDesiderataCmd.Flags().StringVarP(&inputImage, "input-image", "i", "", "input image to add to the desiderata triplet")
	addDesiderataCmd.MarkFlagRequired("input-image")

	addDesiderataCmd.Flags().StringVarP(&outputImage, "output-image", "o", "", "output image to add to the desiderata triplet")
	addDesiderataCmd.MarkFlagRequired("output-image")

	addDesiderataCmd.Flags().StringVarP(&cvmfsRepo, "repository", "r", "", "cvmfs repository add to the desiderata triplet")
	addDesiderataCmd.Flags().StringVarP(&userInput, "user-input", "a", "", "username to access the input registry")
	addDesiderataCmd.Flags().StringVarP(&userOutput, "user-output", "b", "", "username to access the output registry")
	addDesiderataCmd.MarkFlagRequired("repository")

	addDesiderataCmd.Flags().BoolVarP(&convert, "convert", "c", false, "start the conversion process immediately after adding the desiderata")
	addDesiderataCmd.Flags().BoolVarP(&overwriteLayer, "overwrite-layers", "f", false, "overwrite the layer if they are already inside the CVMFS repository")

	rootCmd.AddCommand(addDesiderataCmd)
}

var addDesiderataCmd = &cobra.Command{
	Use:   "add-desiderata",
	Short: "Add a desiderata triplet",
	Run: func(cmd *cobra.Command, args []string) {
		// parse both images
		inputImg, err := lib.ParseImage(inputImage)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		inputImg.User = userInput
		outputImg, err := lib.ParseImage(outputImage)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		outputImg.User = userOutput
		outputImg.IsThin = true

		// check if the images are already in the database
		inputImgDb, errIn := lib.GetImage(inputImg)
		outputImgDb, errOut := lib.GetImage(outputImg)

		// some error, that is not because the image is not in the db
		if errIn != nil && errIn != sql.ErrNoRows {
			lib.LogE(errIn).Fatal("Error in querying the database for the input image")
		}
		if errOut != nil && errOut != sql.ErrNoRows {
			lib.LogE(errOut).Fatal("Error in querying the database for the output image")
		}

		// both images are already in our database
		// check if also the desiderata itself is already in the database
		if errIn == nil && errOut == nil {
			inputId := inputImgDb.Id
			outputId := outputImgDb.Id
			_, err := lib.GetDesiderata(inputId, outputId, cvmfsRepo)
			if err == nil {
				lib.LogE(err).Fatal("Desiderata is already in the database")
			}
		}

		// trying to get the input manifest, if we are not able to get the input image manifest there is something wrong, hence we avoid to add the desiderata to the database itself
		_, err = inputImg.GetManifest()
		if err != nil {
			lib.LogE(err).Fatal("Impossible to get the input manifest")
		}

		// we need to identify the real input and output, if they are already in the db we can use them directly, otherwise we first add them
		var (
			inputImgDbId, outputImgDbId int
		)
		if inputImgDb.Id != 0 {

			inputImgDbId = inputImgDb.Id
		} else {
			err = lib.AddImage(inputImg)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to add the input image to the database")
			}
			inputImgDbId, err = lib.GetImageId(inputImg)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to get the ID of the input image just added")
			}
		}
		if outputImgDb.Id != 0 {
			outputImgDbId = outputImgDb.Id
		} else {
			err = lib.AddImage(outputImg)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to add the out image to the database")
			}
			outputImgDbId, err = lib.GetImageId(outputImg)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to get the ID of the output image just added")
			}

		}
		// at this point we add the real desiderata to the database
		err = lib.AddDesiderata(inputImgDbId, outputImgDbId, cvmfsRepo)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to add the desiderata to the database")
		}

		// if required to convert mmediately the desiderata we do so
		if convert {
			desi, err := lib.GetDesiderataF(inputImgDbId, outputImgDbId, cvmfsRepo)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to retrieve the desiderata just added")
			}
			err = lib.ConvertDesiderata(desi, false, overwriteLayer)
			if err != nil {
				lib.LogE(err).Fatal("Impossible to convert the newly added desiderata")
			}
		}
	},
}
