#!/bin/bash
#
# Test that a plain run of `sonalyze jobs` produces bit-identical output vs the Rust version.
#
# Usage:
#  jobs1.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}
NUMDIFF=${NUMDIFF:-../../numdiff/numdiff}

set -e
$GO_SONALYZE jobs -data-dir "$1" -from "$2" -to "$3" -user - > go-output.txt
$RUST_SONALYZE jobs --data-path "$1" --from "$2" --to "$3" --user - > rust-output.txt
if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
    $NUMDIFF go-output.txt rust-output.txt
fi
rm -f go-output.txt rust-output.txt
