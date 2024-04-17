#!/bin/bash
#
# Like parse6 but includes some more fields with approximately-equal comparisons
#
# Usage:
#  parse6.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}
NUMDIFF=${NUMDIFF:-../../../numdiff/numdiff}

set -e

# We must sort the data here because they are not sorted internally (without any options) and the Go
# and Rust storage layer produce the data in different orders due to concurrency effects.  But for
# json this means we first have to split the data into lines.
#
# We can't print the memory size fields because they are subject to roundoff in the Rust version: it
# rounds to GB on input and loses data.
#
# Printing cpu_util_pct (after -clean) is subject to too many roundoff errors too, although I don't
# know why that should be, so it's probably a bug.

for format in csv csvnamed awk fixed; do
    echo "  $format"
    $GO_SONALYZE parse -data-dir "$1" -from "$2" -to "$3" --fmt $format,all | \
        sort > go-output.txt
    $RUST_SONALYZE parse --data-path "$1" --from "$2" --to "$3" --fmt $format,all | \
        sort > rust-output.txt
    if [[ ! $(cmp go-output.txt rust-output.txt) ]]; then
        $NUMDIFF go-output.txt rust-output.txt
    fi
    rm -f go-output.txt rust-output.txt
done
