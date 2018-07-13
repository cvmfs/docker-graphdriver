package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	rootCmd.AddCommand(downloadManifestCmd)
}

var downloadManifestCmd = &cobra.Command{
	Use:   "download-manifest",
	Short: "Download the manifest of the image, if sucessful it will print the manifest itself, otherwise will show what went wrong.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		img, err := lib.ParseImage(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		url := img.GetManifestUrl()
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
		os.Exit(0)
	},
}
