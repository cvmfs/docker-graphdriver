#!/bin/bash

function setup_graphdriver() {
    local graphdriver_name="$1"
    local plugin_url="$GRAPHDRIVERS_REPO_URL/$graphdriver_name"

    scratch_cleanup
    cd "$SCRATCH"

    mkdir plugin_workdir
    pushd plugin_workdir > /dev/null

    cp /data/config.json .
    sed -i "s/__binary__/$graphdriver_name/" config.json

    mkdir rootfs
    sudo tar xjf "$GRAPH_PLUGIN_ROOTFS_TAR" -C rootfs

    sudo cp "$BINARIES/$graphdriver_name" "rootfs/usr/local/bin/$graphdriver_name"

    popd > /dev/null

    local plugin_name="atlantic777/$graphdriver_name"
    mkdir graph
    sudo dockerd -D --experimental -g "$SCRATCH/graph" -s aufs &>>dockerd.log &

    while ! docker info &>/dev/null; do
        sleep 1
    done

    sudo docker plugin create "$plugin_name" "$SCRATCH/plugin_workdir" > /dev/null
    docker plugin enable "$plugin_name" > /dev/null
    sudo pkill dockerd

    while [ "$(pidof dockerd)" != "" ]; do
        sleep 1
    done
}
