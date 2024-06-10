#!/bin/bash
#
# Script to build things that need to be built.

set -o errexit

echo "======================================================================="
echo " GO SONALYZE RELEASE BUILD + SMOKE TEST"
echo
( cd sonalyze ; go build )
( cd sonalyze ; ./sonalyze help 2&> /dev/null )

echo "======================================================================="
echo " NAICREPORT RELEASE BUILD + SMOKE TEST"
echo
( cd naicreport ; go build )
( cd naicreport ; ./naicreport help 2&> /dev/null )

echo "======================================================================="
echo " SONARD RELEASE BUILD + SMOKE TEST"
echo
( cd sonard ; go build )
( cd sonard ; ./sonard -h 2&> /dev/null )

echo "======================================================================="
echo " SLURMINFO RELEASE BUILD + SMOKE TEST"
echo
( cd slurminfo ; go build )
( cd slurminfo ; ./slurminfo -h 2&> /dev/null )

echo "======================================================================="
echo " MAKE-CLUSTER-CONFIG RELEASE BUILD + SMOKE TEST"
echo
( cd make-cluster-config ; go build )
( cd make-cluster-config ; ./make-cluster-config -h 2&> /dev/null )

echo "======================================================================="
echo " JSONCHECK RELEASE BUILD + SMOKE TEST"
echo
( cd jsoncheck ; go build )
( cd jsoncheck ; ./jsoncheck ../tests/config/good-config.json )

echo "======================================================================="
echo " NUMDIFF RELEASE BUILD + SMOKE TEST"
echo
( cd numdiff ; go build )
( cd numdiff ; ./numdiff numdiff.go numdiff.go )

( cd attic ; ./build.sh )

echo "======================================================================="
echo "======================================================================="
echo "NORMAL COMPLETION"
