#!/bin/bash
#
# Test that a plain run of `sonalyze parse` produces bit-identical output vs the Rust version.
#
# Usage:
#  parse1.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}

set -e
# We must sort the data here because they are not sorted internally (without any options) and the Go
# and Rust storage layer produce the data in different orders due to concurrency effects.
$GO_SONALYZE parse -data-dir "$1" -from "$2" -to "$3" | sort > go-output.txt
$RUST_SONALYZE parse --data-path "$1" --from "$2" --to "$3" | sort > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
