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
echo " RUSTUTILS UNIT TESTS"
echo "======================================================================="
# RUSTUTILS TESTS
( cd rustutils ; cargo test )

echo "======================================================================="
echo " SYSINFO RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sysinfo ; go build )
if [[ $(uname) != Darwin ]]; then
    ( cd sysinfo ; ./sysinfo -h 2&> /dev/null )
fi

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

# No longer compatible, as metadata in the Go version now requires a --bounds argument and I've not
# implemented that in the Rust version.
#
# echo "======================================================================="
# echo " RUST SONALYZE REGRESSION TEST"
# echo "======================================================================="
# ( cd sonalyze ; cargo build )
# ( cd tests ; ./run_tests.sh )

