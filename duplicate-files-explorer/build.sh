#!/bin/bash
# Credits: https://github.com/fmahnke/shell-semver/blob/master/increment_version.sh

version=$(cat "./semver.txt") 
semver_array=( ${version//./ } )

((semver_array[2]++))


echo "${semver_array[0]}.${semver_array[1]}.${semver_array[2]}" > ./semver.txt

echo $(date -u +"%Y-%m-%dT%H:%M:%SZ") > ./buildts.txt

go build -o ../bin/ -race