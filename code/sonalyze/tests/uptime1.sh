#!/bin/bash
#
# Test that a plain run of `sonalyze uptime` produces bit-identical output between versions.  See
# test-generic.sh for info about env vars.

set -e
for opt in "" --only-up --only-down; do
    echo "  $opt"
    $OLD_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 5 $opt > old-output.txt
    $NEW_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 5 $opt > new-output.txt
    cmp old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

