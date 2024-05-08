#!/bin/bash
#
# Test that a run of `sonalyze uptime` with a config file that names a host that's not in the data
# produces bit-identical output vs the Rust version + it has only the single "down" record for the host.
#
# Usage:
#  uptime5.sh data-dir from to

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}

set -e
$GO_SONALYZE uptime -data-dir "$1" -from "$2" -to "$3" -interval 5 -config-file uptime5.cfg > go-output.txt
$RUST_SONALYZE uptime --data-path "$1" --from "$2" --to "$3" --interval 5 --config-file uptime5.cfg > rust-output.txt
cmp go-output.txt rust-output.txt

# TODO: These tests actually belong in the regression test suite

if [[ $(grep -E 'ml5\.hpc\.uio\.no' go-output.txt | wc -l) -ne 1 ]]; then
    echo "Unexpected Go output #1"
    exit 1
fi
if [[ $(grep -E 'ml5\.hpc\.uio\.no.*down' go-output.txt | wc -l) -ne 1 ]]; then
    echo "Unexpected Go output #1"
    exit 1
fi

rm -f go-output.txt rust-output.txt
