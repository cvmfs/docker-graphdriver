package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cvmfs/docker-graphdriver/daemon/lib"
)

var (
	pass, registry string
)

func init() {
	addUserCmd.Flags().StringVarP(&username, "username", "u", "", "username to use to log in into the registry.")
	addUserCmd.Flags().StringVarP(&pass, "password", "p", "", "password to use to log in into the registry.")
	addUserCmd.Flags().StringVarP(&registry, "registry", "r", "", "registry for which use the credential")

	addUserCmd.MarkFlagRequired("username")
	addUserCmd.MarkFlagRequired("password")
	addUserCmd.MarkFlagRequired("registry")
	rootCmd.AddCommand(addUserCmd)
}

var addUserCmd = &cobra.Command{
	Use:   "add-user",
	Short: "Add an user to the database, beware, the password is saved in plain text!",
	Run: func(cmd *cobra.Command, args []string) {
		err := lib.AddUser(username, pass, registry)
		if err != nil {
			lib.LogE(err).Fatal("Problem in adding the user to the database, abort.")
		}
	},
}
