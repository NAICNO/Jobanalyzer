#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze uptime` produces bit-identical output between versions.
#
# Note there is a time dependency here, as the two runs may filter the earliest records differently
# and may have minor differences at the start of the run.  This is a consequence of how eg "20d" is
# interpreted, namely precisely and not as start-of-day of that day.

set -e

source test-settings

for fmt in fixed csv csvnamed awk json; do
    echo "Format old: $fmt,default"
    $OLD_SONALYZE uptime -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -interval 5 -t 1d -fmt ${fmt},device,host,state,start,end > old-output.txt
    $NEW_SONALYZE uptime -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -interval 5 -t 1d -fmt ${fmt},default > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# v0 and v1 default should print the same but the names are different so do only fixed, csv, awk
for fmt in fixed csv awk; do
    echo "Format old vs v1default: $fmt,v1default"
    $OLD_SONALYZE uptime -data-dir "$DATA_PATH" -f 20d -t 1d -config-file "$CONFIG" -interval 5 -fmt $fmt,noheader,device,host,state,start,end > old-output.txt
    $NEW_SONALYZE uptime -data-dir "$DATA_PATH" -f 20d -t 1d -config-file "$CONFIG" -interval 5 -fmt $fmt,noheader,v1default > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# Old and new names should print the same values, ignoring the names
for fmt in fixed csv awk; do
    echo "Format old-vs-new-names: $fmt,all"
    $OLD_SONALYZE uptime -data-dir "$DATA_PATH" -f 20d -t 1d -config-file "$CONFIG" -interval 5 -fmt ${fmt},noheader,device,host,state,start,end > old-output.txt
    $NEW_SONALYZE uptime -data-dir "$DATA_PATH" -f 20d -t 1d -config-file "$CONFIG" -interval 5 -fmt ${fmt},noheader,Device,Hostname,State,Start,End > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

