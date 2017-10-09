#!/bin/bash


function start_overlay() {
	modprobe overlay && PATH=/usr/local/bin:$PATH overlay2_cvmfs
}

function start_aufs() {
	modprobe aufs && PATH=/usr/local/bin:$PATH aufs_cvmfs
}

function fail() {
	echo "ERROR: aufs and overlayfs modules not available!"
}

start_overlay || start_aufs || fail
