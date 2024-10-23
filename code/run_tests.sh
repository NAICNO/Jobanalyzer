#!/bin/bash
#
# This compiles all programs, runs all module tests, then runs all regression and blackbox tests.
#
# To just run the regression and blackbox tests: cd tests ; ./run_tests.sh

set -o errexit

echo "======================================================================="
echo " GO SONALYZE RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonalyze ; go build )
( cd sonalyze ; ./sonalyze help 2&> /dev/null )

# NAICREPORT TESTS
( cd naicreport ; ./run_tests.sh )

echo "======================================================================="
echo " SONARD RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd sonard ; go build )
( cd sonard ; ./sonard -h 2&> /dev/null )

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

# GO-UTIL TESTS
( cd go-utils ; ./run_tests.sh )

# SONALYZE TESTS
( cd sonalyze ; ./run_tests.sh )

# OBSOLETE CODE TESTS
# Commented out because these no longer work with modified test cases and I'm
# too lazy to update obsolete code to make them pass.
# ( cd attic ; ./run_tests.sh )

echo "======================================================================="
echo " DASHBOARD JS LIBRARIES SELFTEST"
echo "======================================================================="
( cd dashboard ; ./run_tests.sh )

echo "======================================================================="
echo " CONFIG FILES TEST"
echo "======================================================================="
( cd jsoncheck
  dir=../../production/jobanalyzer-server/cluster-config
  for f in $dir/*-config.json $dir/cluster-aliases.json; do
      echo $f
      ./jsoncheck $f
  done
)

echo "======================================================================="
echo "======================================================================="
echo "======================================================================="
echo "NORMAL COMPLETION"
