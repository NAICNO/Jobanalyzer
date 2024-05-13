#!/bin/bash
#
# Test that a plain run of `sonalyze profile -fmt=json` produces bit-identical output between
# versions.  See test-generic.sh for info about env vars.
#
# TODO: There's a hack: memory sizes are clamped to zero because there are differences in how those
# values are represented in the two programs, and so they won't be exactly the same.

set -e
old_extra=
new_extra=
# It's a bug in the rust version that a dummy field must be supplied, work around it
if [[ $OLD_NAME == "rust" ]]; then
    old_extra=",cpu"
fi
if [[ $NEW_NAME == "rust" ]]; then
    new_extra=",cpu"
fi
$OLD_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" --fmt json,nomemory${old_extra} | \
    sed 's/},/}\n/g' \
    > old-output.txt
$NEW_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" --fmt json,nomemory${new_extra} | \
    sed 's/},/}\n/g' \
    > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt
