#!/bin/bash
#
# Test that a plain run of `sonalyze load` produces bit-identical output between versions.  See
# test-generic.sh for info about env vars.

set -e
$OLD_SONALYZE load --data-path "$DATA_PATH" --from "$FROM" --to "$TO" > old-output.txt
$NEW_SONALYZE load --data-path "$DATA_PATH" --from "$FROM" --to "$TO" > new-output.txt
if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
    $NUMDIFF old-output.txt new-output.txt
fi
rm -f old-output.txt new-output.txt
