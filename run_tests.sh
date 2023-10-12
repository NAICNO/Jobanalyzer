#!/bin/bash
#
# This compiles all programs, runs all module tests, then runs all regression and blackbox tests.
#
# To just run the regression and blackbox tests: cd tests ; ./run_tests.sh

set -o errexit
( cd sonarlog ; cargo test )
( cd sonalyze ; cargo test ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd naicreport/util ; go test )
( cd naicreport/joblog ; go test )
( cd naicreport/jobstate ; go test )
( cd naicreport/storage ; go test )
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )
( cd tests ; ./run_tests.sh )
