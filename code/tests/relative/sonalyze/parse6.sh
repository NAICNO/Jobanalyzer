#!/bin/bash
#
# Test that a plain run of `sonalyze parse` produces bit-identical output vs an older version, with
# different data formats and with "all" fields.  See test-generic.sh for info about env vars.

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

fields=version,localtime,time,host,cores,user,pid,job,cmd,cpu_pct,gpus,gpu_pct,gpumem_pct,gpu_status,cputime_sec,rolledup
for format in csv csvnamed awk fixed; do
    echo "  $format"
    $OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --fmt $format,$fields | sort > old-output.txt
    $NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --fmt $format,$fields | sort > new-output.txt
    cmp old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

echo "  json"
$OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --fmt json,$fields | \
    sed 's/},/}\n/g;s/\[//g;s/\]//g' | \
    sort > old-output.txt
$NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --fmt json,$fields | \
    sed 's/},/}\n/g;s/\[//g;s/\]//g' | \
    sort > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt

