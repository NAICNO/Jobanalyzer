#!/bin/bash
#
# Like parse6 but includes some more fields with approximately-equal comparisons.  See
# test-generic.sh for info about env vars.

set -e

# We must sort the data here because they are not sorted internally (without any options) and the Go
# and Rust storage layer produce the data in different orders due to concurrency effects.  But for
# json this means we first have to split the data into lines.
#
# We can't print the memory size fields because they are subject to roundoff in the Rust version: it
# rounds to GB on input and loses data.
#
# Printing cpu_util_pct (after -clean) is subject to too many roundoff errors too, although I don't
# know why that should be, so it's probably a bug.

for format in csv csvnamed awk fixed; do
    echo "  $format"
    $OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --fmt $format,all | \
        sort > old-output.txt
    $NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --fmt $format,all | \
        sort > new-output.txt
    if [[ ! $(cmp old-output.txt new-output.txt) ]]; then
        $NUMDIFF old-output.txt new-output.txt
    fi
    rm -f old-output.txt new-output.txt
done
