package main

import (
	"github.com/cvmfs/docker-graphdriver/plugins/aufs_cvmfs/aufs"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/go-plugins-helpers/graphdriver/shim"
)

func main() {
	if print_info() {
		return
	}

	if reexec.Init() {
		return
	}

	h := shim.NewHandlerFromGraphDriver(aufs.Init)
	h.ServeUnix("plugin", 0)
}
