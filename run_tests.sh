#!/bin/bash

function run_test_in_dir() {
        local dir="${1}"
        pushd "${dir}"
        go test --race -coverprofile=coverage.out
        go tool cover -html=coverage.out
        popd
}

export GOTMPDIR="${PWD}/tmp-test-dir/"

mkdir $GOTMPDIR
mkdir "${GOTMPDIR}/heap_test"

test_dirs=("./commons" "./commons/dataStructures" "./duplicate-files-explorer")

for dir in ${test_dirs[@]}; do
        run_test_in_dir "${dir}"
done

rm -fr $GOTMPDIR
unset $GOTMPDIR