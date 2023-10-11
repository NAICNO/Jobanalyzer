#!/bin/bash
#
# Harness for running a bunch of tests that include their own test data.
#
# The way to run this is to first build all the executables (perhaps with the ../test.sh script,
# which actually runs this script) then run this script in its directory.  It will find *.sh in
# specific subdirectories, cd to those directories, and then run those scripts.  Each script runs
# one or more tests and can count on running in its directory.  It will find that $SONALYZE and
# $NAICREPORT points to the appropriate executables.
#
# Each script will do whatever and then pass the name of the test, the expected output, and the
# actual output to the CHECK function.  The latter is defined in prefix.sh in this directory.  That
# file is automatically included as a prefix in every test, and suffix.sh is included as a suffix.
# See eg sonalyze/jobs/aggregate_gpu.sh.
#
# TODO: accept command line arguments with test names to match, for filtering.

export TEST_ROOT=$(pwd)
export SONALYZE=$TEST_ROOT/../sonalyze/target/debug/sonalyze
export NAICREPORT=$TEST_ROOT/../naicreport/naicreport
failed=0
for dir in sonarlog sonalyze format ; do
    for test in $(find $dir -name '*.sh'); do
	export TEST_NAME=$test
	( cd $(dirname $test);
	  bash -c "source $TEST_ROOT/prefix.sh; source $(basename $test); source $TEST_ROOT/suffix.sh "	)
	if (( $? != 0 )); then
	    failed=$((failed+1))
	fi
    done
done
if (( failed > 0 )); then
    echo "FAILED TEST FILES: " $failed
    exit 1
fi
