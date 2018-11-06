package cmd

import (
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/repository-manager/lib"
)

var (
	convertAgain, overwriteLayer, convertSingularity bool
)

func init() {
	convertCmd.Flags().BoolVarP(&overwriteLayer, "overwrite-layers", "f", false, "overwrite the layer if they are already inside the CVMFS repository")
	convertCmd.Flags().BoolVarP(&convertAgain, "convert-again", "g", false, "convert again images that are already successfull converted")
	convertCmd.Flags().BoolVarP(&convertSingularity, "convert-singularity", "s", true, "also create a singularity images")
	rootCmd.AddCommand(convertCmd)
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert the wishes",
	Run: func(cmd *cobra.Command, args []string) {
		AliveMessage()
		defer lib.ExecCommand("docker", "system", "prune", "--force", "--all")

		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			lib.LogE(err).Fatal("Impossible to read the recipe file")
			os.Exit(1)
		}
		recipe, err := lib.ParseYamlRecipeV1(data)
		if err != nil {
			lib.LogE(err).Fatal("Impossible to parse the recipe file")
			os.Exit(1)
		}
		for _, wish := range recipe.Wishes {
			fields := log.Fields{"input image": wish.InputName,
				"repository":   wish.CvmfsRepo,
				"output image": wish.OutputName}
			lib.Log().WithFields(fields).Info("Start conversion of wish")
			err = lib.ConvertWish(wish, convertAgain, overwriteLayer, convertSingularity)
			if err != nil {
				lib.LogE(err).WithFields(fields).Error("Error in converting wish, going on")
			}
		}
	},
}
