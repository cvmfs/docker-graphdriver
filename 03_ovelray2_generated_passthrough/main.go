package main

import (
	"github.com/docker/docker/daemon/graphdriver/overlay2"
	"github.com/docker/docker/pkg/reexec"
	"github.com/docker/go-plugins-helpers/graphdriver/shim"
)

func main() {
	if reexec.Init() {
		return
	}

	h := shim.NewHandlerFromGraphDriver(overlay2.Init)
	h.ServeUnix("plugin", 0)
}
