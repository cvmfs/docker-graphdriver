#!/bin/bash
# start docker with new driver
STORAGE_OPT="cvmfsMountMethod=internal"
sudo dockerd --experimental -D -s "$PLUGIN_NAME" --storage-opt "$STORAGE_OPT" -g graph &
wait_process dockerd up

docker run -itd --name test_1 atlantic777/thin_ubuntu bash &&
docker exec test_1 bash -c "echo hello > /world" &&
docker commit test_1 atlantic777/test &&
docker stop test_1
status_1=$?

docker run -itd --name test_2 atlantic777/test bash &&
docker exec test_2 bash -c "cat /world" &&
docker stop test_2
status_2=$?

sleep 2
stop_docker

if [ $status_1 -ne 0 ]; then
    exit -1
fi

if [ $status_2 -ne 0 ]; then
    exit -1
fi
