package lib

import (
	"fmt"
)

const (
	thinImageVersion = "0.1"
)

var token string

func printUsage() {
	fmt.Println("You need to specify a docker image to download!")
}
