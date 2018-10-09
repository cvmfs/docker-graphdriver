package cmd

import (
	"fmt"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	convertAgain, overwriteLayer bool
)

func init() {
	convertCmd.Flags().BoolVarP(&overwriteLayer, "overwrite-layers", "f", false, "overwrite the layer if they are already inside the CVMFS repository")
	convertCmd.Flags().BoolVarP(&convertAgain, "convert-again", "g", false, "convert again images that are already successfull converted")
	rootCmd.AddCommand(convertCmd)
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert the wishes",
	Run: func(cmd *cobra.Command, args []string) {
		AliveMessage()
		showWeReceivedSignal := make(chan os.Signal, 1)
		signal.Notify(showWeReceivedSignal, os.Interrupt)

		stopWishLoopSignal := make(chan os.Signal, 1)
		signal.Notify(stopWishLoopSignal, os.Interrupt)

		go func() {
			<-showWeReceivedSignal
			lib.Log().Info("Received SIGINT (Ctrl-C) waiting the last layer to upload then exiting.")
		}()

		wish, err := lib.GetAllWishes()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, wish := range wish {
			select {
			case <-stopWishLoopSignal:
				lib.Log().Info("Exiting because of SIGINT")
				os.Exit(1)
			default:
				{
				}
			}
			lib.Log().WithFields(log.Fields{"input image": wish.InputName}).Info("Converting Image")
			err = lib.ConvertWish(wish, convertAgain, overwriteLayer)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		os.Exit(0)
	},
}
