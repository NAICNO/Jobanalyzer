#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze node` produces bit-identical output between versions.

set -e

source test-settings

for fmt in fixed csv csvnamed awk json; do
    echo "Format old: $fmt,default"
    $OLD_SONALYZE node -data-dir "$DATA_PATH" -f 7d -t 1d -fmt ${fmt},default > old-output.txt
    $NEW_SONALYZE node -data-dir "$DATA_PATH"  -f 7d -t 1d -fmt ${fmt},default > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# v0 and v1 default should print the same but the names are different so do only fixed, csv, awk
for fmt in fixed csv awk; do
    echo "Format old vs Default: $fmt,Default"
    $OLD_SONALYZE node -data-dir "$DATA_PATH" -f 7d -t 1d -fmt $fmt,noheader,default > old-output.txt
    $NEW_SONALYZE node -data-dir "$DATA_PATH" -f 7d -t 1d -fmt $fmt,noheader,Default > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# Old and new names should print the same values, ignoring the names
for fmt in fixed csv awk; do
    echo "Format old-vs-new-names: $fmt,all"
    $OLD_SONALYZE node -data-dir "$DATA_PATH" -f 7d -t 1d -fmt ${fmt},noheader,timestamp,host,desc,gpumempct,cores,mem,gpus,gpumem > old-output.txt
    $NEW_SONALYZE node -data-dir "$DATA_PATH" -f 7d -t 1d -fmt ${fmt},noheader,Timestamp,Hostname,Description,GpuMemPct,CpuCores,MemGB,GpuCards,GpuMemGB > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

