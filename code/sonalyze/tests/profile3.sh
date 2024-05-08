#!/bin/bash
#
# Test that a run of `sonalyze profile` (fixed format, all fields except res, presence of nproc
# implied by any process having rolledup > 0) produces bit-identical output vs the Rust version.
#
# Usage:
#  profile3.sh data-dir from to job

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}

set -e
$GO_SONALYZE profile -data-dir "$1" -from "$2" -to "$3" -j "$4" > go-output.txt
$RUST_SONALYZE profile --data-path "$1" --from "$2" --to "$3" -j "$4" > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
