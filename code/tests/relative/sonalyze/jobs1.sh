#!/bin/bash
#
# Test that a plain run of `sonalyze jobs` produces bit-identical output between versions.  See
# test-generic.sh for info about env vars.

set -e
$OLD_SONALYZE jobs --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --user - > old-output.txt
$NEW_SONALYZE jobs --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --user - > new-output.txt
if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
    $NUMDIFF old-output.txt new-output.txt
fi
rm -f old-output.txt new-output.txt
