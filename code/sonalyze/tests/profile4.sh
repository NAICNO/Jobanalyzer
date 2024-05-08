#!/bin/bash
#
# Test that a plain run of `sonalyze profile -fmt=json` produces bit-identical output vs the Rust
# version.  There's a hack: memory sizes are clamped to zero because there are differences in how
# those values are represented in the two programs, and so they won't be exactly the same.
#
# Usage:
#  profile4.sh data-dir from to job
#
# TODO: It's a bug in the rust version that a dummy field must be supplied.

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}

set -e
$GO_SONALYZE profile -data-dir "$1" -from "$2" -to "$3" -j "$4" -fmt json,nomemory | \
    sed 's/},/}\n/g' \
    > go-output.txt
$RUST_SONALYZE profile --data-path "$1" --from "$2" --to "$3" -j "$4" --fmt json,nomemory,cpu | \
    sed 's/},/}\n/g' \
    > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
