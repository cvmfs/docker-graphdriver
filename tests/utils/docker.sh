#!/bin/bash
export DOCKER_BASE_URL="https://get.docker.com/builds/Linux/x86_64"

function docker_tarball_name() {
	  local version="$1"
	  echo "docker-${version}.tgz"
}

function download_docker() {
	  local version="$1"
	  local tarball_name="$(docker_tarball_name $version)"
	  local download_url="${DOCKER_BASE_URL}/${tarball_name}"
	  local target_file="$CACHE/download/$tarball_name"

	  mkdir -p "$CACHE/download"
    wget --quiet "$download_url" -c -O "$target_file"
}

function install_docker() {
    local version="$1"
    local src="$CACHE/download/$(docker_tarball_name $version)"
    local dst="$CACHE/docker_extracted/$version"

    if [ ! -e "$dst" ]; then
        mkdir -p "$dst"
        tar xf "$src" -C "$dst/"
    fi

    sudo cp $dst/docker/docker* /usr/local/bin
}

function setup_docker() {
    local docker_v="$1"

    download_docker "$docker_v"
    install_docker "$docker_v"
}

export -f setup_docker
