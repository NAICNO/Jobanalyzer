#!/bin/bash

set -o errexit

echo "======================================================================="
echo " GO-UTIL UNIT TESTS"
echo "======================================================================="
( cd sonarlog ; go test )
( cd auth ; go test )
( cd time ; go test )
