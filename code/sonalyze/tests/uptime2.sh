#!/bin/bash
#
# Test that a run of `sonalyze uptime -only-up` produces bit-identical output vs the Rust version.
#
# Usage:
#  uptime2.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}

set -e
$GO_SONALYZE uptime -data-dir "$1" -from "$2" -to "$3" -interval 5 -only-up > go-output.txt
$RUST_SONALYZE uptime --data-path "$1" --from "$2" --to "$3" --interval 5 --only-up > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
