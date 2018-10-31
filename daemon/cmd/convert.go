package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
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

		wish, err := lib.GetAllWishes()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, wish := range wish {
			fields := log.Fields{"input image": wish.InputName,
				"repository":   wish.CvmfsRepo,
				"output image": wish.OutputName}
			lib.Log().WithFields(fields).Info("Start conversion of wish")
			err = lib.ConvertWish(wish, convertAgain, overwriteLayer, convertSingularity)
			if err != nil {
				lib.LogE(err).WithFields(fields).Error("Error in converting wish, going on")
			}
		}
		os.Exit(0)
	},
}
