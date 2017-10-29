#!/bin/sh

PLUGINS_DIR="/bin"

function test_unionfs() {
    driver=$1
    cat /proc/filesystems | grep $driver > /dev/null
}

function load_unionfs() {
    driver=$1
    modprobe $driver > /dev/null
}

function run_binary() {
    driver=$1

    echo "runing binary: $driver"

    if [ x"$driver" == x"overlay" ]; then
        exec $PLUGINS_DIR/overlay2_cvmfs
    elif [ x"$driver" == x"aufs" ]; then
        exec $PLUGINS_DIR/aufs_cvmfs
    fi
}

function fail() {
	  echo "ERROR: aufs and overlayfs modules not available!"
    exit 1
}

function start_plugin() {
    driver=$1
    (test_unionfs $driver || load_unionfs $driver) && run_binary $driver
}

start_plugin "overlay" || start_plugin "aufs" || fail
