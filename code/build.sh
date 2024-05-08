#!/bin/bash
#
# Script to build things that need to be built.

set -o errexit

( cd attic ; ./build.sh )

echo "======================================================================="
echo " NAICREPORT RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )

echo "======================================================================="
echo " SONARD RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonard ; go build )
( cd sonard ; ./sonard -h 2&> /dev/null )

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

echo "======================================================================="
echo " MAKE-CLUSTER-CONFIG RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd make-cluster-config ; go build )
( cd make-cluster-config ; ./make-cluster-config -h 2&> /dev/null )

echo "======================================================================="
echo " JSONCHECK RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd jsoncheck ; go build )
( cd jsoncheck ; ./jsoncheck ../tests/config/good-config.json )

echo "======================================================================="
echo " NUMDIFF RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd numdiff ; go build )
( cd numdiff ; ./numdiff numdiff.go numdiff.go )

echo "======================================================================="
echo "======================================================================="
echo "======================================================================="
echo "NORMAL COMPLETION"
