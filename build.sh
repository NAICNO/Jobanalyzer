#!/bin/bash
#
# Script to build things that need to be built.

set -o errexit

echo "======================================================================="
echo " SONALYZE RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyze ; cargo build --release )
( cd sonalyze ; target/release/sonalyze help > /dev/null )
( cd sonalyze ; target/release/sonalyze jobs --fmt=help > /dev/null )

echo "======================================================================="
echo " NAICREPORT RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )

echo "======================================================================="
echo " LOGINFO RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd loginfo ; go build )
( cd loginfo ; ./loginfo help 2&> /dev/null )
