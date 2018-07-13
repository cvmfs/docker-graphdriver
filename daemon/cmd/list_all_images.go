package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	listAllImagesCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(listAllImagesCmd)
}

var listAllImagesCmd = &cobra.Command{
	Use:   "list-images",
	Short: "Show all the images actually in the database",
	Run: func(cmd *cobra.Command, args []string) {
		imgs, err := lib.GetAllImagesInDatabase()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if machineFriendly {
			fmt.Printf("name,scheme,registry,repository,tag,digest,is_thin\n")
			for _, img := range imgs {
				fmt.Printf("%s,%s,%s,%s,%s,%s,%s\n",
					img.WholeName(), img.Scheme,
					img.Registry, img.Repository,
					img.Tag, img.Digest,
					fmt.Sprint(img.IsThin))
			}
		} else {
			for _, img := range imgs {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetAlignment(tablewriter.ALIGN_LEFT)
				table.SetHeader([]string{"Key", "Value"})
				table.Append([]string{"Name", img.WholeName()})
				table.Append([]string{"Scheme", img.Scheme})
				table.Append([]string{"Registry", img.Registry})
				table.Append([]string{"Repository", img.Repository})
				table.Append([]string{"Tag", img.Tag})
				table.Append([]string{"Digest", img.Digest})
				table.Render()
			}
		}
		os.Exit(0)
	},
}
