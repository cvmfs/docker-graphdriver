package cmd

import (
	"fmt"
	"github.com/cvmfs/docker-graphdriver/docker2cvmfs/lib"
	"github.com/spf13/cobra"
	"log"
)

var PrintConfig = &cobra.Command{
	Use:   "config",
	Short: "Print image config",
	Run: func(cmd *cobra.Command, args []string) {
		flag := cmd.Flags().Lookup("registry")
		var registry string = string(flag.Value.String())
		config, err := lib.GetConfig(registry, args[0])
		if err != nil {
			log.Fatal("Impossible to retrieve the configuration")
		}
		fmt.Println(config)
	},
}
