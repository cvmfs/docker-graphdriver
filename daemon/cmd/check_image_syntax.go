package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	machineFriendly bool
)

func init() {
	checkImageSyntaxCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(checkImageSyntaxCmd)
}

var checkImageSyntaxCmd = &cobra.Command{
	Use:   "check-image-syntax",
	Short: "Check that the provide image has a valid syntax, the same checks are applied before any command in the converter.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		img, err := lib.ParseImage(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if machineFriendly {
			fmt.Printf("scheme,registry,repository,tag,digest\n")
			fmt.Printf("%s,%s,%s,%s,%s\n", img.Scheme, img.Registry, img.Repository, img.Tag, img.Digest)
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"Key", "Value"})
			table.Append([]string{"Scheme", img.Scheme})
			table.Append([]string{"Registry", img.Registry})
			table.Append([]string{"Repository", img.Repository})
			table.Append([]string{"Tag", img.Tag})
			table.Append([]string{"Digest", img.Digest})
			table.Render()
		}
		os.Exit(0)
	},
}
