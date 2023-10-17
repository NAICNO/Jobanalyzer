#!/bin/bash
#
# This compiles all programs, runs all module tests, then runs all regression and blackbox tests.
#
# To just run the regression and blackbox tests: cd tests ; ./run_tests.sh

set -o errexit

# Run sonarlog/sonalyze whitebox tests.  I can't get `cargo test --all-features` to work, so test
# separate feature flags separately.
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

# Build sonalyze and test online help
echo "======================================================================="
echo " SONALYZE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd sonalyze ; target/debug/sonalyze jobs --fmt=help > /dev/null )

# Run naicreport whitebox tests
echo "======================================================================="
echo " NAICREPORT UNIT TESTS"
echo "======================================================================="
( cd naicreport/util ; go test )
( cd naicreport/joblog ; go test )
( cd naicreport/jobstate ; go test )
( cd naicreport/storage ; go test )

# Build naicreport and test online help
echo "======================================================================="
echo " NAICREPORT BUILD + SMOKE TEST"
echo "======================================================================="
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )

# Run regression and blackbox tests in default configurations
echo "======================================================================="
echo " SONALYZE REGRESSION TEST, DEFAULT FEATURES"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd tests ; ./run_tests.sh )

# Run regression and blackbox tests in other configurations
echo "======================================================================="
echo " SONALYZE REGRESSION TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonalyze ; cargo build -F untagged_sonar_data )
( cd tests ; ./run_tests.sh )

