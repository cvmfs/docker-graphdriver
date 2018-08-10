package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	listAllWishesCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(listAllWishesCmd)
}

var listAllWishesCmd = &cobra.Command{
	Use:     "list-wishes",
	Aliases: []string{"list-wish", "ls-wish", "ls-wishes", "wishlist", "wish-list", "wishes-list", "wish-ls", "wishes-ls"},
	Short:   "Show all the wishes in the database",
	Run: func(cmd *cobra.Command, args []string) {
		wishes, err := lib.GetAllWishes()
		if err != nil {
			lib.LogE(err).Fatal("Impossible to get the desiderata")
		}
		lib.PrintMultipleWishes(wishes, machineFriendly, true)
		os.Exit(0)
	},
}
