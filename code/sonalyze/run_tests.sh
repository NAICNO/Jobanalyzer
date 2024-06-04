#!/bin/bash

set -e

echo "======================================================================="
echo " GO SONALYZE REGRESSION TEST"
echo "======================================================================="

go build

echo "COMPONENTS"
for i in common db sonarlog; do
    ( cd $i ; go test )
done

echo "TEST SUITE"

export SONALYZE=$(pwd)/sonalyze
( cd ../tests ; ./run_tests.sh )

