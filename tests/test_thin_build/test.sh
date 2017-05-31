#!/bin/bash

BUILD_DIR="$PWD/docker_build_scratch"

mkdir "$BUILD_DIR"
cd "$BUILD_DIR"

cp "$ROOT_DIR/data/Dockerfile" .
echo "Hello world" > test_file

cd "$SCRATCH"

sudo dockerd -D --experimental -g graph -s "$PLUGIN_NAME" &
wait_process dockerd up

docker build "$BUILD_DIR"
status=$?


if $status; then
    docker run atlantic777/build cat /test_file
    $status=$?
fi

stop_docker
exit $status
