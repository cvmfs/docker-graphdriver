#!/bin/bash

set -e

die() {
  echo "$1"
  exit 1
}

[ "x$GOROOT" != x ] || die "GOROOT missing"
[ "x$CVMFS_SOURCE_LOCATION" != x ] || die "CVMFS_SOURCE_LOCATION missing"
[ "x$CVMFS_BUILD_LOCATION" != x ] || die "CVMFS_BUILD_LOCATION missing"

export GOPATH="$CVMFS_SOURCE_LOCATION/.."
PLUGINS_ROOT="github.com/cvmfs/docker-graphdriver/plugins"

cd $CVMFS_BUILD_LOCATION
for plugin in aufs_cvmfs overlay2_cvmfs; do
  echo "Building: $plugin"
  PATH=$GOROOT/bin:$PATH go build -v "$PLUGINS_ROOT/$plugin"
done

