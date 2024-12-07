#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze profile` produces bit-identical output between versions.
#
# There can be (unlikely) spurious failures:
#  - No job found in the preparatory step
#  - Processes are not deterministically sortable because start time, command name, and pid all match
#    (when the pid is 0, which it can be).

set -e

source test-settings

# We want this to pick something with more than one process, so do our best.  It might find nothing.
job=$($OLD_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -min-runtime 1h -u - -fmt awk,job,cmd | grep , | head -n 1 | awk '{ print $1 }')

# No 'default' alias in the old code, so expand it.  Here I add nproc because I don't want to
# test the nproc-insertion logic, as it is different in the new code.
echo "Format old: fixed,default"
$OLD_SONALYZE profile -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -job $job -t 1d -fmt fixed,time,cpu,mem,gpu,gpumem,cmd,nproc > old-output.txt
$NEW_SONALYZE profile -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -job $job -t 1d -fmt fixed,default,nproc > new-output.txt
diff -b old-output.txt new-output.txt
rm -f old-output.txt new-output.txt

# v0 and v1 default (new names) should print the same.  It's broken that this is 'mem' and not 'res' but we're not going to rock that boat.
# Again add nproc to avoid testing the nproc-insertion logic.
echo "Format old vs Default: fixed,default"
$OLD_SONALYZE profile -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -job $job -t 1d -fmt fixed,noheader,time,cpu,mem,gpu,gpumem,cmd,nproc > old-output.txt
$NEW_SONALYZE profile -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -job $job -t 1d -fmt fixed,noheader,Default,nproc > new-output.txt
diff -b old-output.txt new-output.txt
rm -f old-output.txt new-output.txt

# Old and new names should print the same values, ignoring the names.
echo "Format old: fixed,default"
$OLD_SONALYZE profile -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -job $job -t 1d -fmt fixed,noheader,time,cpu,mem,res,gpu,gpumem,cmd,nproc > old-output.txt
$NEW_SONALYZE profile -data-dir "$DATA_PATH" -f 20d -config-file "$CONFIG" -job $job -t 1d -fmt fixed,noheader,Timestamp,CpuUtilPct,VirtualMemGB,ResidentMemGB,Gpu,GpuMemGB,Command,NumProcs > new-output.txt
diff -b old-output.txt new-output.txt
rm -f old-output.txt new-output.txt

# Actual profiles.
for quant in cpu mem res gpu gpumem; do
    for fmt in csv awk; do
        echo "Format old: $fmt,$quant"
        $OLD_SONALYZE profile -data-dir "$DATA_PATH" -f 3d -config-file "$CONFIG" -job $job -t 1d -fmt header,${fmt},${quant} > old-output.txt
        $NEW_SONALYZE profile -data-dir "$DATA_PATH" -f 3d -config-file "$CONFIG" -job $job -t 1d -fmt header,${fmt},${quant} > new-output.txt
        diff -b old-output.txt new-output.txt
        rm -f old-output.txt new-output.txt
    done
done
