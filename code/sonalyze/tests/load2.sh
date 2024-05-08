#!/bin/bash
#
# Test that a run of `sonalyze load` with various options produces bit-identical output vs the Rust version.
#
# Usage:
#  load2.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}
NUMDIFF=${NUMDIFF:-../../numdiff/numdiff}

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
                $GO_SONALYZE load -data-dir "$1" -from "$2" -to "$3" $bucketing $grouping $selection $config > go-output.txt
                $RUST_SONALYZE load --data-path "$1" --from "$2" --to "$3" $bucketing $grouping $selection $config > rust-output.txt
                if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
                    $NUMDIFF go-output.txt rust-output.txt
                fi
                rm -f go-output.txt rust-output.txt
            done
        done
    done
done



