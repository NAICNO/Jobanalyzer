#!/bin/bash
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
