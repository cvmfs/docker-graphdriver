#!/bin/bash

# start docker with new driver
STORAGE_OPT="cvmfsMountMethod=internal"
sudo dockerd --experimental -D -s "$PLUGIN_NAME" --storage-opt "$STORAGE_OPT" -g graph &

while ! docker info &>/dev/null; do
    sleep 1
done

# run the hello world test with a regular image
docker run library/ubuntu:16.04 echo "Hello world"
status1=$?

# run the hello world test with a thin image
docker run atlantic777/thin_ubuntu echo "Hello world"
status2=$?

sleep 2
stop_docker

if [ $status1 -ne 0 ]; then
    exit -1
fi

if [ $status2 -ne 0 ]; then
    exit -1
fi
