#!/bin/bash

function run_test_in_dir() {
        local dir="${1}"
        pushd "${dir}"
        go test
        popd
}

export GOTMPDIR="${PWD}/tmp-test-dir/"

mkdir $GOTMPDIR
mkdir "${GOTMPDIR}/heap_test"

echo "test1" > "${GOTMPDIR}/heap_test/file1"
echo "test1" > "${GOTMPDIR}/heap_test/file2"
echo "test3" > "${GOTMPDIR}/heap_test/file3"

test_dirs=("./commons/ds" "./duplicate-files-explorer")

for dir in ${test_dirs[@]}; do
        run_test_in_dir "${dir}"
done

rm -fr $GOTMPDIR
unset $GOTMPDIR