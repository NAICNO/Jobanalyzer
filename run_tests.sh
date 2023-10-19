#!/bin/bash
#
# This compiles all programs, runs all module tests, then runs all regression and blackbox tests.
#
# To just run the regression and blackbox tests: cd tests ; ./run_tests.sh

set -o errexit

echo "======================================================================="
echo " SONARLOG UNIT TEST, DEFAULT FEATURES"
echo "======================================================================="
( cd sonarlog ; cargo test )

echo "======================================================================="
echo " SONARLOG UNIT TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonarlog ; cargo test -F untagged_sonar_data )

echo "======================================================================="
echo " SONALYZE UNIT TEST, DEFAULT FEATURES"
echo "======================================================================="
( cd sonalyze ; cargo test )

echo "======================================================================="
echo " SONALYZE UNIT TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonalyze ; cargo test -F untagged_sonar_data )

echo "======================================================================="
echo " SONALYZE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd sonalyze ; target/debug/sonalyze jobs --fmt=help > /dev/null )

echo "======================================================================="
echo " NAICREPORT UNIT TESTS"
echo "======================================================================="
( cd naicreport/util ; go test )
( cd naicreport/joblog ; go test )
( cd naicreport/jobstate ; go test )
( cd naicreport/storage ; go test )

echo "======================================================================="
echo " NAICREPORT BUILD + SMOKE TEST"
echo "======================================================================="
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )

echo "======================================================================="
echo " LOGINFO BUILD + SMOKE TEST"
echo "======================================================================="
( cd loginfo ; go build )
( cd loginfo ; ./loginfo help 2&> /dev/null )

echo "======================================================================="
echo " SONALYZE REGRESSION TEST, DEFAULT FEATURES"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd tests ; ./run_tests.sh )

echo "======================================================================="
echo " SONALYZE REGRESSION TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonalyze ; cargo build -F untagged_sonar_data )
( cd tests ; ./run_tests.sh )

