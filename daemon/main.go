package main

import (
	_ "github.com/mattn/go-sqlite3"

	"github.com/cvmfs/docker-graphdriver/daemon/cmd"
)

func main() {
	cmd.EntryPoint()
}
