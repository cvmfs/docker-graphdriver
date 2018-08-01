package cmd

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

func init() {
	listAllUsersCmd.Flags().BoolVarP(&machineFriendly, "machine-friendly", "z", false, "produce machine friendly output, one line of csv")
	rootCmd.AddCommand(listAllUsersCmd)
}

var listAllUsersCmd = &cobra.Command{
	Use:     "list-users",
	Aliases: []string{"list-user", "ls-user", "ls-users", "user-ls", "users-ls"},
	Short:   "List all the user in the database along with the registry they are linked to.",
	Run: func(cmd *cobra.Command, args []string) {
		users, err := lib.GetAllUsers()
		if err != nil {
			lib.LogE(err).Fatal("Problem in retrieving the users from the database.")
		}
		if machineFriendly {
			fmt.Println("user,registry")
			for _, user := range users {
				fmt.Printf("%s,%s\n", user.Username, user.Registry)
			}
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"User", "Registry"})
			for _, user := range users {
				table.Append([]string{user.Username, user.Registry})
			}
			table.Render()
		}
	},
}
