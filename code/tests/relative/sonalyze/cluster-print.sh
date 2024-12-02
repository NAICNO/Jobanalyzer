#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze cluster` produces bit-identical output between versions.
#
# Note, these can fail because the old code (pre-reflection) did not sort alias names, and so the
# old output is a little bit random.

set -e

source test-settings

for fmt in fixed csv csvnamed awk json; do
    echo "Format old: $fmt,default"
    $OLD_SONALYZE cluster -jobanalyzer-dir "$JOBANALYZER_DIR" -fmt ${fmt},default > old-output.txt
    $NEW_SONALYZE cluster -jobanalyzer-dir "$JOBANALYZER_DIR" -fmt ${fmt},default > new-output.txt
    diff old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# v0 and v1 default should print the same but the names are different so do only fixed, csv, awk
for fmt in fixed csv awk; do
    echo "Format old vs Default: $fmt,Default"
    $OLD_SONALYZE cluster -jobanalyzer-dir "$JOBANALYZER_DIR" -fmt $fmt,noheader,default > old-output.txt
    $NEW_SONALYZE cluster -jobanalyzer-dir "$JOBANALYZER_DIR" -fmt $fmt,noheader,Default > new-output.txt
    diff old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# Old and new names should print the same values, ignoring the names
for fmt in fixed csv awk; do
    echo "Format old-vs-new-names: $fmt,all"
    $OLD_SONALYZE cluster -jobanalyzer-dir "$JOBANALYZER_DIR" -fmt ${fmt},noheader,cluster,aliases,desc > old-output.txt
    $NEW_SONALYZE cluster -jobanalyzer-dir "$JOBANALYZER_DIR" -fmt ${fmt},noheader,Name,Aliases,Description > new-output.txt
    diff old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

