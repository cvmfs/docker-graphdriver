package main

import (
	"github.com/cvmfs/docker-graphdriver/plugins/overlay2_cvmfs/overlay2"
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

	h := shim.NewHandlerFromGraphDriver(overlay2.Init)
	h.ServeUnix("plugin", 0)
}
