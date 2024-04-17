#!/bin/bash
#
# Test that a run of `sonalyze parse -merge-by-host-and-job` produces bit-identical output vs the Rust version.
#
# Usage:
#  parse3.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}

set -e
fields=time,user,host,job,cmd
$GO_SONALYZE parse -data-dir "$1" -from "$2" -to "$3" -merge-by-host-and-job -fmt=csvnamed,$fields > go-output.txt
$RUST_SONALYZE parse --data-path "$1" --from "$2" --to "$3" --merge-by-host-and-job --fmt=csvnamed,$fields > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
