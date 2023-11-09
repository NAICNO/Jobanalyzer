#!/bin/bash

set -o errexit

echo "======================================================================="
echo " NAICREPORT UNIT TESTS"
echo "======================================================================="
( cd util ; go test )
( cd joblog ; go test )
( cd jobstate ; go test )
( cd storage ; go test )

echo "======================================================================="
echo " NAICREPORT BUILD + SMOKE TEST"
echo "======================================================================="
( go build )
( ./naicreport help 2&> /dev/null )
