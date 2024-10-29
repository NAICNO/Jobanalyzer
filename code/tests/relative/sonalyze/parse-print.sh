#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze parse` produces bit-identical output between versions.

set -e

source test-settings

for fmt in fixed csv csvnamed awk; do
    echo "Format old: $fmt,default"
    $OLD_SONALYZE parse -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,job,user,cmd | sort > old-output.txt
    $NEW_SONALYZE parse -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,default | sort > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

for fmt in fixed csv awk; do
    echo "Format old: $fmt,v1default"
    $OLD_SONALYZE parse -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,noheader,job,user,cmd | sort > old-output.txt
    $NEW_SONALYZE parse -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,noheader,v1default | sort > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# The "all" fields are not quite comparable (MB vs KB, ppid) but we can print most of them
# csvnamed won't work, of course

for fmt in fixed csv awk; do
    echo "Format old: $fmt,noheader,many-fields"
    $OLD_SONALYZE parse -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,noheader,version,localtime,host,cores,user,pid,job,cmd,cpu_pct,gpus,gpu_pct,gpumem_pct,gpu_status,cputime_sec,rolledup,cpu_util_pct | sort > old-output.txt
    $NEW_SONALYZE parse -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,noheader,version,Timestamp,Host,Cores,User,Pid,Job,Cmd,CpuPct,Gpus,GpuPct,GpuMemPct,GpuFail,CpuTimeSec,Rolledup,CpuUtilPct | sort > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

