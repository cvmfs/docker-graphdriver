#!/bin/bash

IMG_DIR="$PWD/docker_thin_scratch"
BUILD_DIR="$PWD/docker_build_scratch"

mkdir "$BUILD_DIR" "$IMG_DIR"

cd "$BUILD_DIR"
cp "$ROOT_DIR/data/thin_scratch/Dockerfile" .
echo "Hello world" > "test_file"

cd "$IMG_DIR"
cp "$ROOT_DIR/data/thin_scratch/thin_scratch" ".thin"

cd "$SCRATCH"

sudo dockerd -D --experimental -g graph -s "$PLUGIN_NAME" &
wait_process dockerd up

tar c "image" | docker imoprt - "thin_scratch"
docker build "$BUILD_DIR"

status=$?

stop_docker
exit $status
