package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(removeAllConverted)
}

var removeAllConverted = &cobra.Command{
	Use:     "remove-converted",
	Aliases: []string{"rm-converted", "rm-convert", "converted-rm", "convert-rm"},
	Short:   "Remove all the reference to already converted images, basically reset the status of the database",
	Run: func(cmd *cobra.Command, args []string) {
		n, err := lib.DeleteAllConverted()
		if err != nil {
			lib.LogE(err).Fatal("Error in removing the converted references")
		}
		lib.Log().WithFields(log.Fields{"removed references": n}).Info("Removed all the reference to converted images")
	},
}
