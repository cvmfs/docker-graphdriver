#!/bin/bash
REPO_NAME="docker_graphdriver_plugins"
REPO="github.com/atlantic777/$REPO_NAME"
WORKSPACE="/tmp/workspace/$REPO_NAME"

go get  "$REPO/06_aufs_cvmfs/..."
go test "$REPO/06_aufs_cvmfs/..."
status=$?

if [ "$status" == "0" ]; then
    echo "Unit tests passed"
    exit 0
else
    echo "Unit tests failed!"
    exit -1
fi
