package main

import (
	"github.com/cvmfs/docker-graphdriver/docker2cvmfs/cmd"
)

func main() {
	if print_info() {
		return
	}

	cmd.RootCmd.Execute()
}
