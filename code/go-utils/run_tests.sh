#!/bin/bash

set -o errexit

echo "======================================================================="
echo " GO-UTIL UNIT TESTS"
echo "======================================================================="
( cd alias ; go test )
( cd auth ; go test )
( cd config ; go test )
( cd filesys ; go test )
( cd freecsv ; go test )
( cd gpuset ; go test )
( cd hostglob ; go test )
( cd sysinfo ; go test )
( cd time ; go test )
if [[ $(uname) != Darwin ]]; then
    ( cd sysinfo ; go test )
fi
( cd time ; go test )
