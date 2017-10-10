package main

import (
	"fmt"
	"os"
)

var (
	version  = "<unofficial build>"
	git_hash = "<unofficial build>"
	help_msg = "This program is part of the CernVM File System\n" +
		"It serves the Docker graphdriver plugins HTTP API on port 80.\n" +
		"The plugin binary is meant to be started in a dedicated Docker plugin container.\n\n" +
		"Refer to https://github.com/cvmfs/docker-graphdriver for further details."
)

func print_info() bool {
	if len(os.Args) < 2 {
		return false
	}

	switch arg := os.Args[1]; arg {
	case "-v":
		fmt.Println("Version: ", version)
		fmt.Println("Commit:  ", git_hash)
		return true

	case "-h":
		fmt.Println(help_msg)
		return true
	}
	return false
}
