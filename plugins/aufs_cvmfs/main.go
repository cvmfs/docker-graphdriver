package main

import (
	"github.com/cvmfs/docker-graphdriver/plugins/aufs_cvmfs/aufs"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/go-plugins-helpers/graphdriver/shim"

	"fmt"
	"os"
)

var (
	version  = "<unofficial build>"
	git_hash = "<unofficial build>"
	help_msg = "This program serves the Docker graphdriver plugins HTTP API on port 80.\n" +
		"The plugin binary is meant to be started in a dedicated Docker plugin container.\n\n" +

		"Refer to https://github.com/cvmfs/docker-graphdriver for further details."
)

func main() {
	if len(os.Args) > 1 {
		print_info()
		return
	}

	if reexec.Init() {
		return
	}

	h := shim.NewHandlerFromGraphDriver(aufs.Init)
	h.ServeUnix("plugin", 0)
}

func print_info() {
	if len(os.Args) > 0 {
		switch arg := os.Args[1]; arg {
		case "-v":
			fmt.Println("Version: ", version)
			fmt.Println("Commit:  ", git_hash)
		case "-h":
			fmt.Println(help_msg)
		}
		return
	}
}
