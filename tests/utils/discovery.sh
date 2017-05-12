#!/bin/bash
. "$TESTS/common.sh"

function collect_tests() {
    local filter="$1"

    if [ "$filter" != "" ]; then
        test_list=`ls $TESTS | grep "^test_" | grep "$filter"`
    else
        test_list=`ls $TESTS | grep "^test_"`
    fi
}

function execute_tests() {
    for t in $test_list
    do
        output="$(bash -x $TESTS/$t/test.sh 2>&1)"
        status=$?

        if [ $status -ne 0 ]; then
            echo "Test failed: $(basename $t)"
            echo "$output"
            return -1
        else
            echo "Test passed: $(basename $t)"
        fi
    done

    return 0
}

function run_test_suite() {
    local filter="$1"

    collect_tests "$filter"
    execute_tests

    return $?
}
