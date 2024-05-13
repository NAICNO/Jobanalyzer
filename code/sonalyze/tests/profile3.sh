#!/bin/bash
#
# Test that a run of `sonalyze profile` (fixed format, all fields except res, presence of nproc
# implied by any process having rolledup > 0) produces bit-identical output between versions.  See
# test-generic.sh for info about env vars.

set -e
$OLD_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" > old-output.txt
$NEW_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt
