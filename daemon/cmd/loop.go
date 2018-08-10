package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	loopCmd.Flags().BoolVarP(&overwriteLayer, "overwrite-layers", "f", false, "overwrite the layer if they are already inside the CVMFS repository")
	loopCmd.Flags().BoolVarP(&convertAgain, "convert-again", "g", false, "convert again images that are already successfull converted")
	rootCmd.AddCommand(loopCmd)
}

var loopCmd = &cobra.Command{
	Use:   "loop",
	Short: "An infinite loop that keep converting all the images",
	Run: func(cmd *cobra.Command, args []string) {
		for {
			wish, err := lib.GetAllWishes()
			if err != nil {
				lib.LogE(err).Error("Error in getting the desiderata")
			}
			for _, wish := range wish {
				fields := log.Fields{
					"input image":  wish.InputName,
					"CVMFS repo":   wish.CvmfsRepo,
					"output image": wish.OutputName,
				}
				lib.Log().WithFields(fields).Info("Working on desiderata")
				err = lib.ConvertWish(wish, convertAgain, overwriteLayer)
				if err != nil {
					lib.LogE(err).Error("Error in converting the desiderata")
				}
			}
		}
	},
}
