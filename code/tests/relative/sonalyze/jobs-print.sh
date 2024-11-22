#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze jobs` produces bit-identical output between versions.

set -e

source test-settings

# All fields with old names, while we're moving things
# This will fail occasionally as "now" changes between the two runs

for fmt in fixed csv csvnamed awk; do
    echo "Format old: $fmt,all"
    $OLD_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,jobm,job,user,duration,duration/sec,start,start/sec,end,end/sec,cpu-avg,cpu-peak,rcpu-avg,rcpu-peak,mem-avg,mem-peak,rmem-avg,rmem-peak,res-avg,res-peak,rres-avg,rres-peak,gpu-avg,gpu-peak,rgpu-avg,rgpu-peak,sgpu-avg,sgpu-peak,gpumem-avg,gpumem-peak,rgpumem-avg,rgpumem-peak,sgpumem-avg,sgpumem-peak,gpus,gpufail,cmd,host,now,now/sec,classification,cputime/sec,cputime,gputime/sec,gputime > old-output.txt
    $NEW_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,all > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# Default fields, ditto

for fmt in fixed csv csvnamed awk; do
    echo "Format old: $fmt,default"
    $OLD_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,std,cpu,mem,gpu,gpumem,cmd > old-output.txt
    $NEW_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,default > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# New names
for fmt in csv awk; do
    echo "Format old vs new"
    $OLD_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,noheader,jobm,job,user,duration,duration/sec,start,start/sec,end,end/sec,cpu-avg,cpu-peak,rcpu-avg,rcpu-peak,mem-avg,mem-peak,rmem-avg,rmem-peak,res-avg,res-peak,rres-avg,rres-peak,gpu-avg,gpu-peak,rgpu-avg,rgpu-peak,sgpu-avg,sgpu-peak,gpumem-avg,gpumem-peak,rgpumem-avg,rgpumem-peak,sgpumem-avg,sgpumem-peak,gpus,gpufail,cmd,host,now,now/sec,classification,cputime/sec,cputime,gputime/sec,gputime > old-output.txt
    $NEW_SONALYZE jobs -data-dir "$DATA_PATH" -config-file "$CONFIG" -f 2d -t 1d -fmt $fmt,noheader,JobAndMark,Job,User,Duration,Duration/sec,Start,Start/sec,End,End/sec,CpuAvgPct,CpuPeakPct,RelativeCpuAvgPct,RelativeCpuPeakPct,MemAvgGB,MemPeakGB,RelativeMemAvgPct,RelativeMemPeakPct,ResidentMemAvgGB,ResidentMemPeakGB,RelativeResidentMemAvgPct,RelativeResidentMemPeakPct,GpuAvgPct,GpuPeakPct,RelativeGpuAvgPct,RelativeGpuPeakPct,OccupiedRelativeGpuAvgPct,OccupiedRelativeGpuPeakPct,GpuMemAvgGB,GpuMemPeakGB,RelativeGpuMemAvgPct,RelativeGpuMemPeakPct,OccupiedRelativeGpuMemAvgPct,OccupiedRelativeGpuMemPeakPct,Gpus,GpuFail,Cmd,Host,Now,Now/sec,Classification,CpuTime/sec,CpuTime,GpuTime/sec,GpuTime > new-output.txt
    diff -b old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

