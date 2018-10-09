package cmd

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(setRecipeCmd)
}

// We start by converting the YamlRecipe in a set of images, that we add to the catalog of images.
var setRecipeCmd = &cobra.Command{
	Use:   "set-recipe",
	Short: "It changes the whish list to match the recipe, indepotent action",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			lib.LogE(err).Fatal("Impossible to read the recipe file")
		}

		actualWishes, err := lib.GetAllWish()
		if err != nil {
			lib.LogE(err).Fatal("Impossible to get all the wishes from the db")
		}
		fmt.Println(actualWishes)
		recipe, err := lib.ParseYamlRecipeV1(data)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to parse the recipe file")
		}

		toAddWish := []lib.Wish{}
		toRemoveWish := []lib.Wish{}

		for _, newWish := range recipe.Wishes {
			alreadyPresent := false
			for _, oldWish := range actualWishes {
				if newWish.Equal(oldWish) {
					alreadyPresent = true
					break
				}
			}
			if !alreadyPresent {
				toAddWish = append(toAddWish, newWish)
			}
		}
		for _, oldWish := range actualWishes {
			toKeep := false
			for _, newWish := range recipe.Wishes {
				if oldWish.Equal(newWish) {
					toKeep = true
					break
				}
			}
			if toKeep == false {
				toRemoveWish = append(toRemoveWish, oldWish)
			}
		}

		fmt.Println(recipe)
		fmt.Println(toAddWish)
		fmt.Println(toRemoveWish)
		for _, newWish := range toAddWish {
			err = lib.AddWish(newWish.InputImage, newWish.OutputImage, newWish.CvmfsRepo)
			if err != nil {
				input, _ := lib.GetImageById(newWish.InputImage)
				output, _ := lib.GetImageById(newWish.OutputImage)
				lib.LogE(err).WithFields(log.Fields{"input image": input.WholeName(), "repo": newWish.CvmfsRepo, "output image": output.WholeName()}).Warning("Error in adding a wish to the database")
			}
		}

		for _, oldWish := range toRemoveWish {
			input, _ := lib.GetImageById(oldWish.InputImage)
			output, _ := lib.GetImageById(oldWish.OutputImage)
			n, err := lib.DeleteWish(oldWish.Id)
			if err != nil {
				lib.LogE(err).WithFields(log.Fields{"input image": input.WholeName(), "repo": oldWish.CvmfsRepo, "output image": output.WholeName()}).Warning("Error in removing a wish")
			}
			if n > 1 {
				lib.LogE(err).Warning("Remove more than one line from the database while removing a wish, should not happen")
			}
		}

	},
}
