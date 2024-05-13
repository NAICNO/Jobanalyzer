#!/bin/bash
#
# Test that a plain run of `sonalyze uptime` with an unusually short `-interval` produces
# bit-identical output between versions.  See test-generic.sh for info about env vars.

set -e
$OLD_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 2 > old-output.txt
$NEW_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 2 > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt
