#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze sacct` produces bit-identical output between versions.

set -e

source test-settings

# This works but not for json.  Neither old nor new sacct printer sorts the output in any reasonable sense.
# So I skip JSON here, on faith that it'll work if the others work.
for fmt in fixed csv csvnamed awk; do
    echo "Format old: $fmt,default"
    $OLD_SONALYZE sacct -data-dir "$SACCT_DATA_PATH" -config-file "$SACCT_CONFIG" -f 2d -t 1d -fmt ${fmt},default | sort > old-output.txt
    $NEW_SONALYZE sacct -data-dir "$SACCT_DATA_PATH" -config-file "$SACCT_CONFIG" -f 2d -t 1d -fmt ${fmt},default | sort > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# for sacct, 'rmem' is actually resident memory not virtual memory like everywhere else, so this test is still valid
for fmt in fixed csv awk; do
    echo "Format old vs Default: $fmt,Default"
    $OLD_SONALYZE sacct -data-dir "$SACCT_DATA_PATH" -config-file "$SACCT_CONFIG" -f 2d -t 1d -fmt $fmt,noheader,default | sort > old-output.txt
    $NEW_SONALYZE sacct -data-dir "$SACCT_DATA_PATH" -config-file "$SACCT_CONFIG" -f 2d -t 1d -fmt $fmt,noheader,Default | sort > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# Mostly names are the same for the two.  Old and new names should print the same values, ignoring the names
for fmt in fixed csv awk; do
    echo "Format old-vs-new-names: $fmt,all"
    $OLD_SONALYZE sacct -data-dir "$SACCT_DATA_PATH" -config-file "$SACCT_CONFIG" -f 2d -t 1d -fmt ${fmt},noheader,Start,End,Submit,RequestedCPU,UsedCPU,rcpu,rmem,User,JobName,State,Account,Reservation,Layout,NodeList,JobID,MaxRSS,ReqMem,ReqCPUS,ReqGPUS,ReqNodes,Elapsed,Suspended,Timelimit,ExitCode,Wait,Partition,ArrayJobID,ArrayIndex | sort > old-output.txt
    $NEW_SONALYZE sacct -data-dir "$SACCT_DATA_PATH" -config-file "$SACCT_CONFIG" -f 2d -t 1d -fmt ${fmt},noheader,Start,End,Submit,RequestedCPU,UsedCPU,RelativeCPU,RelativeResidentMem,User,JobName,State,Account,Reservation,Layout,NodeList,JobID,MaxRSS,ReqMem,ReqCPUS,ReqGPUS,ReqNodes,Elapsed,Suspended,Timelimit,ExitCode,Wait,Partition,ArrayJobID,ArrayIndex | sort > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

