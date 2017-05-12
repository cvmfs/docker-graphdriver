#!/bin/bash
export TEST_SCRIPTS_DIR="$PWD/tests"
export ROOT_DIR="$PWD"
export WORKSPACE="$PWD/workspace"
mkdir workspace

cd "$WORKSPACE"

filter="$1"

. $TEST_SCRIPTS_DIR/common.sh

if [ "$filter" != "" ]; then
    test_list=`ls $TEST_SCRIPTS_DIR | grep "^test_" | grep "$filter"`
else
    test_list=`ls $TEST_SCRIPTS_DIR | grep "^test_"`
fi

for t in $test_list
do
    output="$(bash -x $TEST_SCRIPTS_DIR/$t/test.sh 2>&1)"
    status=$?

    if [ $status -ne 0 ]; then
        echo "Test failed: $(basename $t)"
        echo "$output"
        exit -1
    else
        echo "Test passed: $(basename $t)"
    fi
done
