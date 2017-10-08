#!/bin/bash

set -e

die() {
  echo "$1"
  exit 1
}

[ "x$CVMFS_SOURCE_LOCATION" != x ] || die "CVMFS_SOURCE_LOCATION missing"
[ "x$CVMFS_BUILD_LOCATION" != x ] || die "CVMFS_BUILD_LOCATION missing"

GIT_COMMIT=$(cd $CVMFS_SOURCE_LOCATION && git rev-parse HEAD)

cd $CVMFS_BUILD_LOCATION
make -f $CVMFS_SOURCE_LOCATION/ci/container/Makefile

