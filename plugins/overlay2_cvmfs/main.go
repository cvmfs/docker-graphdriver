package main

import (
	"github.com/cvmfs/docker-graphdriver/plugins/overlay2_cvmfs/overlay2"
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

	h := shim.NewHandlerFromGraphDriver(overlay2.Init)
	h.ServeUnix("plugin", 0)
}
