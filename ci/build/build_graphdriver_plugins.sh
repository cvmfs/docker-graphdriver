#!/bin/bash
PLUGINS_REPO="github.com/atlantic777/docker_graphdriver_plugins"
PLUGINS_REPO_PATH="$GOPATH/src/$PLUGINS_REPO"

go get -v "$PLUGINS_REPO/..."

mkdir -p binaries
rm -rf binaries/*

pushd binaries

for i in `ls $PLUGINS_REPO_PATH | grep "^[0-9]"`
do
   echo "Build: $i"
   go build -v "$PLUGINS_REPO/$i"

   if [ ! -e `basename $i` ]; then
       echo "Build failed!"
       ls
       exit -1
   fi
   echo "------------------------------------"
done

popd
