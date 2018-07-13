package main

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/cvmfs/docker-graphdriver/daemon/cmd"
)

var (
	InsertDesiderata *sql.Stmt
)

/*

Three simple commands:
	1. add desiderata
	2. refresh
	3. check

add desiderata will add a new row to the desiderata table and will try to convert it.
refresh will check all the converted and be sure that they are up to date, maybe we changed a tag
check will simply make sure that every desiderata is been converted

We will talk with the docker daemon and use it for most of the work

*/

var (
	username_input, password_input,
	input_repository,

	username_output, password_output,
	output_repository,
	cvmfs_repository string

	machineFriendly bool
)

func main() {
	cmd.EntryPoint()
}

func SplitReference(repository string) (register, tag, digest string, err error) {
	tryTag := strings.Split(repository, ":")
	if len(tryTag) == 2 {
		return tryTag[0], tryTag[1], "", nil
	}
	tryDigest := strings.Split(repository, "@")
	if len(tryDigest) == 2 {
		return tryTag[0], "", tryTag[1], nil
	}
	return "", "", "", fmt.Errorf("Impossible to split the repository into repository and tag")
}
