#!/bin/bash

set -o errexit

echo "======================================================================="
echo " GO-UTIL UNIT TESTS"
echo "======================================================================="
( cd auth ; go test )
( cd config ; go test )
( cd filesys ; go test )
( cd freecsv ; go test )
( cd hostglob ; go test )
( cd sonarlog ; go test )
( cd time ; go test )
