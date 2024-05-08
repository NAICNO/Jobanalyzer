#!/bin/bash
#
# Test that a plain run of `sonalyze profile -fmt=csv` produces bit-identical output vs the Rust version.
#
# Usage:
#  profile1.sh data-dir from to job

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}

set -e
for format in csv html fixed; do
    echo "  $format"
    for field in cpu mem gpu res gpumem; do
        echo "    $field"
        for max in 0 50 200; do
            echo "      $max"
            # Go and Rust differ for 0: for Go it means "no argument present" and for Rust it means
            # that the argument is actually 0.  Work around this.
            maxarg=""
            if [[ $max != 0 ]]; then
                maxarg="--max $max"
            fi
            $GO_SONALYZE profile -data-dir "$1" -from "$2" -to "$3" -j "$4" $maxarg -fmt csv,$field > go-output.txt
            $RUST_SONALYZE profile --data-path "$1" --from "$2" --to "$3"  -j "$4" $maxarg --fmt csv,$field > rust-output.txt
            cmp go-output.txt rust-output.txt
            rm -f go-output.txt rust-output.txt
        done
    done
done


