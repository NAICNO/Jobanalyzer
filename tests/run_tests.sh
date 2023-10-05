#!/bin/bash
#
# Harness for running a bunch of tests that include their own test data.
#
# The way to run this is to first build all the executables (perhaps with the ../test.sh script,
# which actually runs this script) then run this script in its directory.  It will find *.sh in
# specific subdirectories, cd to those directories, and then run those scripts.  Each script runs
# one or more tests and can count on running in its directory.  It will find that SONALYZE and
# NAICREPORT points to the appropriate executables.
#
# Each script will define the variables t_name (the name of the test), t_expected (the literal
# expected string output) and t_output (the ditto actual output), and then include the file
# harness.sh in the present directory.  That file will run code to test that the output and the
# expected output match.  See eg sonalyze/jobs/aggregate_gpu.sh.

export SONALYZE=$(pwd)/../sonalyze/target/debug/sonalyze
export NAICREPORT=$(pwd)/../naicreport/naicreport
for dir in sonarlog sonalyze ; do
    for test in $(find $dir -name '*.sh'); do
	( cd $(dirname $test); ./$(basename $test) )
    done
done
