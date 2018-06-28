#!/bin/sh

PLUGINS_DIR="/usr/bin"
DRIVER_GUARD="/driver"

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

    if [ -f $DRIVER_GUARD ]; then
      if [ x"$(cat $DRIVER_GUARD)" != x"$driver" ]; then
        echo "This plugin was used with driver $(cat $DRIVER_GUARD) before."
        echo "Cannot switch to $driver, exiting"
        return 1
      fi
    fi

    echo "runing binary: $driver"
    echo "$driver" > $DRIVER_GUARD

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
