#!/bin/bash
REPO_NAME="docker-graphdriver"
REPO="github.com/cvmfs/$REPO_NAME"
WORKSPACE="/tmp/workspace/$REPO_NAME"
GOPATH="$WORKSPACE/cache/gopath"

SRC="plugins"
DST="$GOPATH/src/$REPO/plugins"

mkdir -p "$DST" > /dev/null
cp -r "$SRC"/* "$DST"

go get  "$REPO/plugins/..."
go test "$REPO/plugins/..."

status=$?

if [ "$status" == "0" ]; then
    echo "Unit tests passed"
    exit 0
else
    echo "Unit tests failed!"
    exit -1
fi
