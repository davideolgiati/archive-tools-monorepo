#!/bin/bash

function run_test_in_dir() {
        local dir="${1}"
        pushd "${dir}"
        go test
        popd
}

export GOTMPDIR="${PWD}/tmp-test-dir/"

mkdir $GOTMPDIR

test_dirs=("./commons/ds")

for dir in ${test_dirs[@]}; do
        run_test_in_dir "${dir}"
done

rm -fr $GOTMPDIR
unset $GOTMPDIR