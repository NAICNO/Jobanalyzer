#!/bin/bash
#
# Test that `sonalyze profile -bucket n` produces bit-identical output between versions.  See
# test-generic.sh for info about env vars.

set -e
$OLD_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" --bucket 5 --fmt csv,cpu > old-output.txt
$NEW_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" --bucket 5 --fmt csv,cpu > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt
