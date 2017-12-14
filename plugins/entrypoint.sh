#!/bin/sh

PLUGINS_DIR="/usr/bin"

test_unionfs() {
    driver=$1
    cat /proc/filesystems | grep $driver > /dev/null
}

load_unionfs() {
    driver=$1
    modprobe $driver > /dev/null
}

run_binary() {
    driver=$1

    echo "runing binary: $driver"

    if [ x"$driver" == x"overlay" ]; then
        exec $PLUGINS_DIR/overlay2_cvmfs
    elif [ x"$driver" == x"aufs" ]; then
        exec $PLUGINS_DIR/aufs_cvmfs
    fi
}

fail() {
	  echo "ERROR: aufs and overlayfs modules not available!"
    exit 1
}

start_plugin() {
    driver=$1
    (test_unionfs $driver || load_unionfs $driver) && run_binary $driver
}

init_config() {
    cp -a /cvmfs_ext_config/* /etc/cvmfs/
}

init_config
if uname -r | grep cernvm; then
  start_plugin "aufs" || fail
fi
start_plugin "overlay" || start_plugin "aufs" || fail
