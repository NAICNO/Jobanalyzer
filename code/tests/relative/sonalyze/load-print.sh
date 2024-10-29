#!/bin/bash
#
# test-settings should be a link to a file with settings, see eg settings-naic-monitor.uio.no
#
# Test that a `sonalyze load` produces bit-identical output between versions.

set -e

source test-settings

for fmt in fixed csv csvnamed awk json; do
    # date,time,cpu,mem,gpu,gpumem,gpumask are default fields but `default` was introduced with the
    # reflective formatter.
    echo "Format old: $fmt,default"
    $OLD_SONALYZE load --data-path "$DATA_PATH" --from 3d --to 1d -fmt ${fmt},date,time,cpu,mem,gpu,gpumem,gpumask > old-output.txt
    $NEW_SONALYZE load --data-path "$DATA_PATH" --from 3d --to 1d -fmt ${fmt},default > new-output.txt
    if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
        $NUMDIFF old-output.txt new-output.txt
    fi
    rm -f old-output.txt new-output.txt
done

# v0 and v1 default should print the same but the names are different so do only fixed, csv, awk
for fmt in fixed csv awk; do
    echo "Format old vs v1default: $fmt,v1default"
    $OLD_SONALYZE load --data-path "$DATA_PATH" --from 3d --to 1d -fmt $fmt,noheader,date,time,cpu,mem,gpu,gpumem,gpumask > old-output.txt
    $NEW_SONALYZE load --data-path "$DATA_PATH" --from 3d --to 1d -fmt $fmt,noheader,v1default > new-output.txt
    if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
        $NUMDIFF old-output.txt new-output.txt
    fi
done

# All fields (except "now"), but no alias for this yet
for fmt in fixed csv csvnamed awk json; do
    echo "Format old-names: $fmt,all"
    $OLD_SONALYZE load -data-dir "$DATA_PATH" -config-file $CONFIG -from 3d -to 1d -fmt ${fmt},datetime,date,time,cpu,rcpu,mem,rmem,res,rres,gpu,rgpu,gpumem,rgpumem,gpus,host > old-output.txt
    $NEW_SONALYZE load -data-dir "$DATA_PATH" -config-file $CONFIG -from 3d -to 1d -fmt ${fmt},datetime,date,time,cpu,rcpu,mem,rmem,res,rres,gpu,rgpu,gpumem,rgpumem,gpus,host > new-output.txt
    if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
        $NUMDIFF old-output.txt new-output.txt
    fi
    rm -f old-output.txt new-output.txt
done

# Old and new names should print the same values, ignoring the names
for fmt in fixed csv awk; do
    echo "Format old-vs-new-names: $fmt,all"
    $OLD_SONALYZE load -data-dir "$DATA_PATH" -config-file $CONFIG -from 3d -to 1d -fmt ${fmt},noheader,datetime,date,time,cpu,rcpu,mem,rmem,res,rres,gpu,rgpu,gpumem,rgpumem,gpus,host > old-output.txt
    $NEW_SONALYZE load -data-dir "$DATA_PATH" -config-file $CONFIG -from 3d -to 1d -fmt ${fmt},noheader,DateTime,Date,Time,Cpu,RelativeCpu,VirtualGB,RelativeVirtualMem,ResidentGB,RelativeResidentMem,Gpu,RelativeGpu,GpuGB,RelativeGpuMem,Gpus,Hostname > new-output.txt
    if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
        $NUMDIFF old-output.txt new-output.txt
    fi
    rm -f old-output.txt new-output.txt
done

