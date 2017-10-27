package main

import (
	"fmt"
	"os"
)

var (
	version  = "<unofficial build>"
	git_hash = "<unofficial build>"
	help_msg = "This program is part of the CernVM File System\n" +
		"The docker2cvmfs utility extracts layers from an image in a\n" +
		"Docker registry and creates the thin image (descriptor).\n\n" +
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
