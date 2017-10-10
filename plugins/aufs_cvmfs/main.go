package main

import (
	"github.com/cvmfs/docker-graphdriver/plugins/aufs_cvmfs/aufs"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/go-plugins-helpers/graphdriver/shim"

	"os"
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
