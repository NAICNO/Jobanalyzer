#!/bin/bash
#
# Test that a run of `sonalyze parse` with a cleaning/merging option produces bit-identical output
# between versions.  See test-generic.sh for info about env vars.

set -e
fields=time,user,host,job,cmd
for howto in --clean --merge-by-host-and-job --merge-by-job; do
    echo "  $howto"
    $OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $howto --fmt=csvnamed,$fields > old-output.txt
    $NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $howto --fmt=csvnamed,$fields > new-output.txt
    cmp old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

