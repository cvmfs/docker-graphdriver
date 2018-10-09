package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(dbLocationCmd)
}

var dbLocationCmd = &cobra.Command{
	Use:     "db-location",
	Short:   "Print the location of the database",
	Aliases: []string{"where-db"},
	Run: func(cmd *cobra.Command, args []string) {
		absPath, err := filepath.Abs(lib.DatabaseLocation)
		if err != nil {
			lib.LogE(err).Warning("Error in getting the absolute path of the database")
		}
		fmt.Println(absPath)
	},
}
