#!/bin/bash
export DOCKER_VERSIONS=(
    "17.04.0-ce"
    "17.03.1-ce"
    "17.03.0-ce"
    "1.13.1"
    # "1.13.0"
)

export GRAPHDRIVERS=(
    "06_aufs_cvmfs"
    "07_overlay2_cvmfs"
)

function init() {
  export ROOT_DIR="$(git rev-parse --show-toplevel)"
  export BINARIES="$ROOT_DIR/binaries"

  export TESTS="$ROOT_DIR/tests"
	export WORKSPACE="$ROOT_DIR/workspace"

  export CACHE="$WORKSPACE/cache"
  export SCRATCH="$WORKSPACE/scratch"

  export GOPATH="$CACHE/gopath"

  export GRAPHDRIVERS_REPO_URL="github.com/atlantic777/docker_graphdriver_plugins"
  export GRAPH_PLUGIN_ROOTFS_TAR="/data/ubuntu_cvmfs-2.4.x_rootfs.tar.bz2"

	mkdir -p "$CACHE" "$SCRATCH" "$GOPATH"

  . "$TESTS/utils/docker.sh"
  . "$TESTS/utils/graph.sh"
  . "$TESTS/utils/discovery.sh"

}

function destroy() {
  sudo rm -rf $SCRATCH/*
  sudo rm -rf /usr/local/bin/*
}

function scratch_cleanup() {
    sudo rm -rf $SCRATCH/*
}

# run tests
# - normal
# - paranoid
# - full
function run_tests() {
  local filter="$1"
  status=0

	for docker_v in ${DOCKER_VERSIONS[@]}
	do
    echo "Using docker: $docker_v"

    scratch_cleanup
    setup_docker "$docker_v"

		for graphdriver_plugin in "${GRAPHDRIVERS[@]}"
		do
      export PLUGIN_NAME="atlantic777/$graphdriver_plugin"
      echo "Using plugin: $PLUGIN_NAME"

      setup_graphdriver "$graphdriver_plugin"
      run_test_suite "$filter"
      let "status += $?"
		done
	done
}

filter="$1"
init
run_tests "$filter"
destroy

if [ "x$status" != "x0" ]
then
    exit -1
else
    exit 0
fi
