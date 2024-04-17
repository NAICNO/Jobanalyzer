#!/bin/bash

set -e
go build
( cd common ; go test )
( cd sonarlog ; go test )
( cd tests ; ./run_tests.sh )
