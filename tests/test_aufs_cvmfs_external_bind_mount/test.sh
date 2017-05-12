#!/bin/bash
STORAGE_OPT="cvmfsMountMethod=external"
sudo dockerd --experimental -D -s "$PLUGIN_NAME" --storage-opt "$STORAGE_OPT" -g graph &
wait_process dockerd up

docker run library/ubuntu:16.04 echo "Hello world"
status1=$?

PLUGIN_ID=$(docker plugin inspect --format="{{.Id}}" $PLUGIN_NAME)
CVMFS_MOUNT_PATH="$PWD/graph/plugins/$PLUGIN_ID/rootfs/mnt/$PLUGIN_NAME/cvmfs/docker2cvmfs-ci.cern.ch"
sudo mkdir -p "$CVMFS_MOUNT_PATH"
sudo mount --bind /cvmfs/docker2cvmfs-ci.cern.ch "$CVMFS_MOUNT_PATH"

docker run atlantic777/thin_ubuntu echo "Hello world"
status2=$?

sleep 2
sudo umount "$CVMFS_MOUNT_PATH"

sleep 2
stop_docker

if [ $status1 -ne 0 ]; then
    exit -1
fi

if [ $status2 -ne 0 ]; then
    exit -1
fi
