#!/bin/bash
#
# Test that a plain run of `sonalyze load` produces bit-identical output vs the Rust version.
#
# Usage:
#  load1.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}
NUMDIFF=${NUMDIFF:-../../../numdiff/numdiff}

set -e
$GO_SONALYZE load -data-dir "$1" -from "$2" -to "$3" > go-output.txt
$RUST_SONALYZE load --data-path "$1" --from "$2" --to "$3" > rust-output.txt
if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
    $NUMDIFF go-output.txt rust-output.txt
fi
rm -f go-output.txt rust-output.txt
