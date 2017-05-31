#!/bin/bash

sudo dockerd -D --experimental -g graph -s "$PLUGIN_NAME" &
wait_process dockerd up

docker login "$DOCKERHUB_URL" -u cernvm -p cernvm
login_status=$?

SRC_IMAGE="atlantic777/thin_ubuntu"
DST_IMAGE="$DOCKERHUB_URL/thin_ubuntu-$RANDOM"

if [ "x$login_status" == "x0" ]; then
    docker pull "$SRC_IMAGE"
    docker tag "$SRC_IMAGE" "$DST_IMAGE"
    docker push "$DST_IMAGE"
    status=$?
else
    status=$login_status
fi

stop_docker
exit $status
