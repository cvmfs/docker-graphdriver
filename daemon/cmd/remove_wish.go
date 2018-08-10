package cmd

import (
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(removeWish)
}

var removeWish = &cobra.Command{
	Use:   "remove-wish",
	Short: "Remove a wish from the database starting by id",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			narg, err := strconv.Atoi(arg)
			if err != nil {
				lib.LogE(err).WithFields(log.Fields{"arg": arg}).Warning("Impossible to convert the argument to an integer, please provide as input the id of the wish you want to remove")
				continue
			}
			n, err := lib.DeleteWish(narg)
			if err != nil {
				lib.LogE(err).Error("Error in removing the wish")
			}
			if n != 1 {
				lib.Log().Warning("Remove more than one line in the database")
			}
		}
	},
}
