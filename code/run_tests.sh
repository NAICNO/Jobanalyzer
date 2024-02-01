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
( cd sonarlog ; cargo test -F "untagged_sonar_data" )

echo "======================================================================="
echo " SONALYZE UNIT TEST, NO FEATURES"
echo "======================================================================="
( cd sonalyze ; cargo test --no-default-features -F "" )

echo "======================================================================="
echo " SONALYZE UNIT TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonalyze ; cargo test --no-default-features -F "untagged_sonar_data" )

echo "======================================================================="
echo " SONALYZE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd sonalyze ; target/debug/sonalyze jobs --fmt=help > /dev/null )

# NAICREPORT TESTS
( cd naicreport ; ./run_tests.sh )

echo "======================================================================="
echo " SONARD RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonard ; go build )
( cd sonard ; ./sonard -h 2&> /dev/null )

echo "======================================================================="
echo " SYSINFO RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sysinfo ; go build )
( cd sysinfo ; ./sysinfo -h 2&> /dev/null )

echo "======================================================================="
echo " EXFILTRATE RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd exfiltrate ; go build )
( cd exfiltrate ; ./exfiltrate -h 2&> /dev/null )

echo "======================================================================="
echo " INFILTRATE RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd infiltrate ; go build )
( cd infiltrate ; ./infiltrate -h 2&> /dev/null )

echo "======================================================================="
echo " SONALYZED RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyzed ; go build )
( cd sonalyzed ; ./sonalyzed -h 2&> /dev/null )

echo "======================================================================="
echo " SLURMINFO RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd slurminfo ; go build )
( cd slurminfo ; ./slurminfo -h 2&> /dev/null )

# GO-UTIL TESTS
( cd go-utils ; ./run_tests.sh )

echo "======================================================================="
echo " SONALYZE REGRESSION TEST, NO FEATURES"
echo "======================================================================="
( cd sonalyze ; cargo build --no-default-features -F "" )
( cd tests ; ./run_tests.sh )

echo "======================================================================="
echo " SONALYZE REGRESSION TEST, FEATURE: UNTAGGED DATA"
echo "======================================================================="
( cd sonalyze ; cargo build --no-default-features -F "untagged_sonar_data" )
( cd tests ; ./run_tests.sh )

echo "======================================================================="
echo "======================================================================="
echo "======================================================================="
echo "NORMAL COMPLETION"
