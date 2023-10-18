#!/bin/bash
#
# Harness for running a bunch of tests that include their own test data.
#
# Usage:
#   run_tests.sh [pattern]
#
# The way to run this is to first build all the executables (perhaps with the ../run_tests.sh
# script, which actually runs this script) then run this script in its directory.  It will find *.sh
# in specific subdirectories, cd to those directories, and then run those scripts with prefix and
# suffix scripts loaded along with them, see below.  Each script runs one or more tests and can
# count on running in its directory.
#
# This script defines the following variables that the test can depend on:
#
#   TEST_ROOT - the root directory of the test suite
#   TEST_NAME - the file name of the test being run
#   SONALYZE - the name of the sonalyze executable, typically a debug build
#   NAICREPORT - the name of the naicreport executable
#
# Each script will do whatever and then pass the name of the test (interpreted in the context of
# TEST_NAME), the expected output, and the actual output to the CHECK function.  The latter is
# defined in prefix.sh in this directory.  When all the tests have run, code in the suffix.sh file
# in this directory will compute the necessary error codes and signal those.
#
# See any test file in any subdirectory for examples.
#
# The pattern is a regex pattern that must match the name of the test filname.

TEST_DIRECTORIES="sonarlog sonalyze naicreport"

export TEST_ROOT=$(pwd)
export SONALYZE=$TEST_ROOT/../sonalyze/target/debug/sonalyze
export NAICREPORT=$TEST_ROOT/../naicreport/naicreport

pattern="$1"
hard_failed=0
soft_failed=0
for dir in $TEST_DIRECTORIES ; do
    for test in $(find $dir -name '*.sh'); do
        if [[ $pattern != "" && !($test =~ $pattern) ]]; then
            continue
        fi
	export TEST_NAME=$test
	( cd $(dirname $test);
	  bash -c "source $TEST_ROOT/prefix.sh; source $(basename $test); source $TEST_ROOT/suffix.sh "	)
	exitcode=$?
	if (( exitcode != 0 )); then
	    if (( exitcode == 2 )); then
		soft_failed=$((soft_failed+1))
	    else
		hard_failed=$((hard_failed+1))
	    fi
	fi
    done
done

# Exit with an error code only if there were any hard errors.

if (( soft_failed > 0 || hard_failed > 0 )); then
    echo "FAILED TEST FILES: " $((soft_failed + hard_failed))
    echo "  TOTAL FILES WITH HARD ERRORS: " $hard_failed
    echo "  TOTAL FILES WITH SOFT ERRORS: " $soft_failed
    if ((hard_failed > 0)); then
	exit 1
    fi
fi
echo ""
echo "SUCCESS.  Regression tests succeeded with no unknown failures."
