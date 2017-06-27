#!/bin/bash

function die {
    status=$1
    stop_docker
    exit $status
}

IMG_DIR="$PWD/image"
BUILD_DIR="$PWD/build"

IMAGE_NAME="test-image"

mkdir "$BUILD_DIR" "$IMG_DIR"

cd "$BUILD_DIR"
cp "$GRAPH_PLUGIN_ROOTFS_TAR" .
cp $ROOT_DIR/data/thin_dockerfile_add_tarball/* .


cd "$IMG_DIR"
cp "$ROOT_DIR/data/thin_scratch/thin_scratch" "$IMG_DIR/.thin"

cd "$SCRATCH"

sudo dockerd -D --experimental -g graph -s "$PLUGIN_NAME" &
wait_process dockerd up

tar c "$IMG_DIR" | docker import - "thin_scratch"        || die $?
docker build "$BUILD_DIR" -t "$IMAGE_NAME"               || die $?
docker run "$IMAGE_NAME" echo "Hello world"              || die $?
docker run "$IMAGE_NAME" cat /hello                      || die $?

stop_docker
exit 0
