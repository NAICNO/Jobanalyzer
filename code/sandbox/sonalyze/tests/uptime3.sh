#!/bin/bash
#
# Test that a run of `sonalyze uptime -only-down` produces bit-identical output vs the Rust version.
#
# Usage:
#  uptime3.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}

set -e
$GO_SONALYZE uptime -data-dir "$1" -from "$2" -to "$3" -interval 5 -only-down > go-output.txt
$RUST_SONALYZE uptime --data-path "$1" --from "$2" --to "$3" --interval 5 --only-down > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
