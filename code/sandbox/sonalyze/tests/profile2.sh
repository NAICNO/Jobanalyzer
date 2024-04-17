#!/bin/bash
#
# Test that `sonalyze profile -bucket n` produces bit-identical output vs the Rust version.
#
# Usage:
#  profile2.sh data-dir from to job

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}

set -e
$GO_SONALYZE profile -data-dir "$1" -from "$2" -to "$3" -j "$4" -bucket 5 -fmt csv,cpu > go-output.txt
$RUST_SONALYZE profile --data-path "$1" --from "$2" --to "$3"  -j "$4" --bucket 5 --fmt csv,cpu > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
