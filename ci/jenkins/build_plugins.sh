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
GIT_COMMIT=$(cd $CVMFS_SOURCE_LOCATION/$PLUGINS_ROOT && git rev-parse HEAD)

cd $CVMFS_BUILD_LOCATION
for plugin in aufs_cvmfs overlay2_cvmfs; do
  echo "Building: $plugin"

  VERSION=$(cat $CVMFS_SOURCE_LOCATION/$PLUGINS_ROOT/$plugin/VERSION)
  PATH=$GOROOT/bin:$PATH go build \
    -ldflags="-X main.version=$VERSION -X main.git_hash=$GIT_COMMIT" \
    -v "$PLUGINS_ROOT/$plugin"
done

