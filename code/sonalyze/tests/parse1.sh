#!/bin/bash
#
# Test that a run of `sonalyze parse` produces bit-identical output between versions.  See
# test-generic.sh for info about env vars.

set -e
# We must sort the data here because they are not sorted internally and different versions produce
# the data in different orders due to concurrency effects.
$OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" | sort > old-output.txt
$NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" | sort > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt
