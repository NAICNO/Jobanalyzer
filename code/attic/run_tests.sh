#!/bin/bash
#
# Test cases for obsolete things.

set -o errexit

echo "======================================================================="
echo " RUST SONARLOG UNIT TEST, DEFAULT FEATURES"
echo "======================================================================="
( cd sonarlog ; cargo test )

echo "======================================================================="
echo " RUST SONARLOG UNIT TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonarlog ; cargo test -F "untagged_sonar_data" )

echo "======================================================================="
echo " RUST SONALYZE UNIT TEST"
echo "======================================================================="
( cd sonalyze ; cargo test )

echo "======================================================================="
echo " RUST SONALYZE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd sonalyze ; target/debug/sonalyze jobs --fmt=help > /dev/null )

echo "======================================================================="
echo " RUST SONALYZE REGRESSION TEST"
echo "======================================================================="
( cd sonalyze ; cargo build )
export SONALYZE=$(pwd)/sonalyze/target/debug/sonalyze
( cd ../tests ; ./run_tests.sh )

echo "======================================================================="
echo " RUSTUTILS UNIT TESTS"
echo "======================================================================="
# RUSTUTILS TESTS
( cd rustutils ; cargo test )

