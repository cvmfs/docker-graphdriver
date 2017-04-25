#!/bin/bash
if [ x"$GOPATH" == x"" ]; then
	echo "GOPATH is not set!"
	exit -1
fi

PLUGINS_REPO="github.com/atlantic777/docker_graphdriver_plugins"
PLUGINS_REPO_PATH="$GOPATH/src/$PLUGINS_REPO"

mkdir -p "$PLUGINS_REPO_PATH" > /dev/null
cp -r plugins/* "$PLUGINS_REPO_PATH"

mkdir -p binaries > /dev/null
rm -rf binaries/*

go get -v $PLUGINS_REPO/...

pushd binaries > /dev/null

for i in `ls $PLUGINS_REPO_PATH | grep "cvmfs"`
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
