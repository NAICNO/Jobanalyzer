#!/bin/bash
#
# Test that a run of `sonalyze load` with various options produces bit-identical output between
# versions.  See test-generic.sh for info about env vars.

set -e
allfields=now,datetime,date,time,cpu,rcpu,mem,rmem,res,rres,gpu,rgpu,gpumem,rgpumem,gpus,host
for selection in "" --all --last --compact; do
    for grouping in "" --group; do
        for config in "" "--config-file $CONFIG --fmt $allfields"; do
            for bucketing in --half-hourly --hourly --half-daily --daily --weekly --none; do
                if [[ $grouping == "--group" && $bucketing == "--none" ]]; then
                    continue
                fi
                echo "  $selection $grouping $bucketing $config"
                $OLD_SONALYZE load --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $bucketing $grouping $selection $config \
                              > old-output.txt
                $NEW_SONALYZE load --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $bucketing $grouping $selection $config \
                              > new-output.txt
                if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
                    $NUMDIFF old-output.txt new-output.txt
                fi
                rm -f old-output.txt new-output.txt
            done
        done
    done
done



