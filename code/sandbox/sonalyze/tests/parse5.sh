#!/bin/bash
#
# Test that a run of `sonalyze parse` with various record filters produces bit-identical output vs
# the Rust version.
#
# Usage:
#  parse5.sh data-dir from to

DATADIR=$1
FROMDATE=$2
TODATE=$3
TESTDATAFILE=$4
GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../../sonalyze/target/release/sonalyze}

# The $TESTDATAFILE must provide a script that initializes the value array appropriate for the test
# data available.  See parse5-*.sh for an example.

declare -A value
source $TESTDATAFILE

expand_opts () {
    local option=$1
    local s=""
    for v in ${value[$option]}; do
        s="$s $option $v"
    done
    echo $s
}

set -e

# We must sort the data here because they are not sorted internally (without any options) and the Go
# and Rust storage layer produce the data in different orders due to concurrency effects.

for i in ${!value[@]}; do
    echo "  $i"
    $GO_SONALYZE parse -data-dir "$DATADIR" -from "$FROMDATE" -to "$TODATE" $opt $(expand_opts $i) \
        | sort > go-output.txt
    $RUST_SONALYZE parse --data-path "$DATADIR" --from "$FROMDATE" --to "$TODATE" $opt $(expand_opts $i) \
        | sort > rust-output.txt
    cmp go-output.txt rust-output.txt
    rm -f go-output.txt rust-output.txt
done

# Special case for bool arguments

$GO_SONALYZE parse -data-dir "$DATADIR" -from "$FROMDATE" -to "$TODATE" $opt --exclude-system-jobs \
    | sort > go-output.txt
$RUST_SONALYZE parse --data-path "$DATADIR" --from "$FROMDATE" --to "$TODATE" $opt --exclude-system-jobs \
    | sort > rust-output.txt
cmp go-output.txt rust-output.txt
rm -f go-output.txt rust-output.txt
