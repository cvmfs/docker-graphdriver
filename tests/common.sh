#!/bin/bash

function prepare_plugin
{
    plugin_bin=$1
    rootfs_tar=$2

    # create workdir
    mkdir plugin_workdir
    pushd plugin_workdir

    # create config.json
    cp /data/config.json .
    sed -i "s/__binary__/$PLUGIN_NAME/" config.json

    # create unpack rootfs
    mkdir rootfs
    sudo tar xjf "$ROOTFS_TAR" -C rootfs

    # copy binary
    sudo cp -v "$PLUGIN_BIN" rootfs/usr/local/bin
    popd
}

function install_plugin
{
    mkdir graph
    sudo dockerd --log-level=fatal -s aufs -g graph & #&>/dev/null &

    while [ "$(pidof dockerd)" == "" ]
    do
        sleep 1
        echo "."
    done

    # create & enable plugin
    sudo docker plugin create "atlantic777/$PLUGIN_NAME" plugin_workdir
    docker plugin enable "atlantic777/$PLUGIN_NAME"

    sudo rm -rf plugin_workdir

    # stop daemon
    sudo pkill dockerd

    while [ "$(pidof dockerd)" != "" ]
    do
        sleep 1
        echo "."
    done
}

function wait_process
{
    process=$1
    target_status=$2

    if [ "$target_status" == "up" ]; then
        while [ "$(pidof $process)" == "" ]
        do
            sleep 1
            echo "."
        done
    elif [ "$target_status" == "down" ]; then
        while [ "$(pidof $process)" != "" ]
        do
            sleep 1
            echo "."
        done
    fi
}

function start_docker
{
    graph_driver=$1

    if [ "$graph_driver" == "" ]; then
        graph_driver="aufs"
    fi

    if [ ! -e graph ]; then
        mkdir -p graph
    fi

    sudo dockerd --experimental -D -s $graph_driver -g graph & #  &>/dev/null &
    wait_process dockerd up
}

function stop_docker
{
    sudo pkill dockerd
    wait_process dockerd down
}

function hello_world_test
{
    PLUGIN_BIN=$1
    ROOTFS_TAR=$2

    prepare_plugin $PLUGIN_BIN $ROOTFS_TAR
    start_docker aufs

    # create & enable plugin
    sudo docker plugin create "atlantic777/$PLUGIN_NAME" plugin_workdir
    docker plugin enable "atlantic777/$PLUGIN_NAME"

    # switch to new graphdriver
    stop_docker
    start_docker "atlantic777/$PLUGIN_NAME"

    # run the echo test
    docker run ubuntu:16.04 echo "Hello world!"
    status=$?

    # cleanup
    stop_docker
    sudo rm -rf plugin_workdir
    sudo rm -rf graph

    return $status
}

export -f prepare_plugin
export -f install_plugin
export -f start_docker
export -f stop_docker
export -f wait_process
export -f hello_world_test
