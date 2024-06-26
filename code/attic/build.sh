#!/bin/bash
#
# Script to build mostly obsolete things that it is still useful to build.

set -o errexit
echo "======================================================================="
echo " ATTIC: RUST SONALYZE DEBUG+RELEASE BUILD + SMOKE TEST"
echo
( cd sonalyze ; cargo build )
( cd sonalyze ; target/debug/sonalyze help > /dev/null )
( cd sonalyze ; target/debug/sonalyze jobs --fmt=help > /dev/null )

( cd sonalyze ; cargo build --release )
( cd sonalyze ; target/release/sonalyze help > /dev/null )
( cd sonalyze ; target/release/sonalyze jobs --fmt=help > /dev/null )

echo "======================================================================="
echo " ATTIC: SYSINFO RELEASE BUILD + SMOKE TEST"
echo
( cd sysinfo ; go build )
if [[ $(uname) != Darwin ]]; then
    ( cd sysinfo ; ./sysinfo -h 2&> /dev/null )
fi

echo "======================================================================="
echo " ATTIC: EXFILTRATE RELEASE BUILD + SMOKE TEST"
echo
( cd exfiltrate ; go build )
( cd exfiltrate ; ./exfiltrate -h 2&> /dev/null )

echo "======================================================================="
echo " ATTIC: INFILTRATE RELEASE BUILD + SMOKE TEST"
echo
( cd infiltrate ; go build )
( cd infiltrate ; ./infiltrate -h 2&> /dev/null )

echo "======================================================================="
echo " ATTIC: SONALYZED RELEASE BUILD + SMOKE TEST"
echo
( cd sonalyzed ; go build )
( cd sonalyzed ; ./sonalyzed -h 2&> /dev/null )

