#!/bin/bash
#
# Test that a run of `sonalyze uptime` with a config file that names a host that's not in the data
# produces bit-identical output between versions + it has only the single "down" record for the
# host.

set -e
$OLD_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 5 --config-file $UPTIME5_CONFIG > old-output.txt
$NEW_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 5 --config-file $UPTIME5_CONFIG > new-output.txt
cmp old-output.txt new-output.txt

# TODO: These tests actually belong in the regression test suite

if [[ $(grep -E "$UPTIME5_HOST" new-output.txt | wc -l) -ne 1 ]]; then
    echo "Unexpected output #1"
    exit 1
fi
if [[ $(grep -E "$UPTIME5_HOST.*down" new-output.txt | wc -l) -ne 1 ]]; then
    echo "Unexpected output #2"
    exit 1
fi

rm -f old-output.txt new-output.txt
