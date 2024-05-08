#!/bin/bash
#
# Test that a run of `sonalyze jobs` with an aggregation filter produces bit-identical output vs the
# Rust version.  We always supply a config file.
#
# Usage:
#  jobs2.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}
NUMDIFF=${NUMDIFF:-../../numdiff/numdiff}

declare -A filter
filter["--min-samples"]=10
filter["--min-runtime"]=10m
filter["--min-cpu-avg"]=200
filter["--min-cpu-peak"]=300
filter["--max-cpu-avg"]=5000
filter["--max-cpu-peak"]=8000
filter["--min-mem-avg"]=20
filter["--min-mem-peak"]=40
filter["--min-res-avg"]=20
filter["--min-res-peak"]=40
filter["--min-gpu-avg"]=10
filter["--min-gpu-peak"]=60
filter["--min-gpumem-avg"]=5
filter["--min-gpumem-peak"]=10

declare -A rfilter
rfilter["--min-rcpu-avg"]=30
rfilter["--min-rcpu-peak"]=40
rfilter["--max-rcpu-avg"]=30
rfilter["--max-rcpu-peak"]=40
rfilter["--min-rmem-avg"]=30
rfilter["--min-rmem-peak"]=40
rfilter["--min-rres-avg"]=30
rfilter["--min-rres-peak"]=40
rfilter["--min-rgpu-avg"]=10
rfilter["--min-rgpu-peak"]=20
rfilter["--max-rgpu-avg"]=50
rfilter["--max-rgpu-peak"]=60
rfilter["--min-rgpumem-avg"]=10
rfilter["--min-rgpumem-peak"]=20

bfilters="--no-gpu --some-gpu --completed --running --zombie --batch"

fmt=std,cpu,mem,res,gpu,gpumem,cmd
rfmt=std,cpu,rcpu,mem,rmem,res,rres,gpu,rgpu,gpumem,rgpumem,cmd

set -e
for num in "" "--numjobs 5"; do
    echo "  numjobs: $num"
    for f in ${!filter[@]}; do
        v=${filter[$f]}
        echo "  $f = $v"
        $GO_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - --fmt $fmt $f $v $num > go-output.txt
        $RUST_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - --fmt $fmt $f $v $num > rust-output.txt
        #wc -l go-output.txt rust-output.txt
        if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
            $NUMDIFF go-output.txt rust-output.txt
        fi
        rm -f go-output.txt rust-output.txt
    done

    for f in ${!rfilter[@]}; do
        v=${rfilter[$f]}
        echo "  $f = $v"
        $GO_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - --fmt $fmt --config-file $CONFIG $f $v $num > go-output.txt
        $RUST_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - --fmt $fmt --config-file $CONFIG $f $v $num > rust-output.txt
        #wc -l go-output.txt rust-output.txt
        if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
            $NUMDIFF go-output.txt rust-output.txt
        fi
        rm -f go-output.txt rust-output.txt
    done

    for b in $bfilters; do
        echo "  $b"
        $GO_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - --fmt $fmt --config-file $CONFIG $b $num > go-output.txt
        $RUST_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - --fmt $fmt --config-file $CONFIG $b $num > rust-output.txt
        #wc -l go-output.txt rust-output.txt
        if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
            $NUMDIFF go-output.txt rust-output.txt
        fi
        rm -f go-output.txt rust-output.txt
    done
done
