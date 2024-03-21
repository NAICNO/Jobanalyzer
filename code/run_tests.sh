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
echo " SONALYZE UNIT TEST"
echo "======================================================================="
( cd sonalyze ; cargo test )

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
echo " JSONCHECK RELEASE BUILD + SMOKE TEST"
echo "======================================================================="
( cd jsoncheck ; go build )
( cd jsoncheck ; ./jsoncheck ../tests/config/good-config.json )

# GO-UTIL TESTS
( cd go-utils ; ./run_tests.sh )

echo "======================================================================="
echo " RUSTUTILS UNIT TESTS"
echo "======================================================================="
# RUSTUTILS TESTS
( cd rustutils ; cargo test )

echo "======================================================================="
echo " SONALYZE REGRESSION TEST"
echo "======================================================================="
( cd sonalyze ; cargo build )
( cd tests ; ./run_tests.sh )

echo "======================================================================="
echo " DASHBOARD JS LIBRARIES SELFTEST"
echo "======================================================================="
( cd dashboard ; ./run_tests.sh )

echo "======================================================================="
echo " CONFIG FILES TEST"
echo "======================================================================="
( cd jsoncheck ; ./jsoncheck ../../production/jobanalyzer-server/scripts/mlx.hpc.uio.no/mlx.hpc.uio.no-config.json )
( cd jsoncheck ; ./jsoncheck ../../production/jobanalyzer-server/scripts/fox.educloud.no/fox.educloud.no-config.json )
( cd jsoncheck ; ./jsoncheck ../../production/jobanalyzer-server/scripts/saga.sigma2.no/saga.sigma2.no-config.json )
( cd jsoncheck ; ./jsoncheck ../../production/jobanalyzer-server/cluster-aliases.json )

echo "======================================================================="
echo "======================================================================="
echo "======================================================================="
echo "NORMAL COMPLETION"
