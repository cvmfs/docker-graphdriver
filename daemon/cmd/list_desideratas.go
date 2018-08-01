package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	listAllDesiderataCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(listAllDesiderataCmd)
}

var listAllDesiderataCmd = &cobra.Command{
	Use:   "list-desideratas",
	Short: "Show all the desiderata in the database",
	Run: func(cmd *cobra.Command, args []string) {
		desideratas, err := lib.GetAllDesiderata()
		if err != nil {
			lib.LogE(err).Fatal("Impossible to get the desiderata")
		}
		lib.PrintMultipleDesideratas(desideratas, machineFriendly, true)
		os.Exit(0)
	},
}
