#!/bin/bash
#
# Test that a run of `sonalyze parse` with various record filters produces bit-identical output vs
# an old version.  See test-generic.sh for info about env vars.

# PARSE5_FILTER is an associative array mapping option name to option value for various record
# filters.

set -e

expand_opts () {
    local option=$1
    local s=""
    for v in ${PARSE5_FILTER[$option]}; do
        s="$s $option $v"
    done
    echo $s
}

# We must sort the data here because they are not sorted internally (without any options) and the Go
# and Rust storage layers produce the data in different orders due to concurrency effects.

for i in ${!PARSE5_FILTER[@]}; do
    echo "  $i"
    $OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $opt $(expand_opts $i) \
        | sort > old-output.txt
    $NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $opt $(expand_opts $i) \
        | sort > new-output.txt
    cmp old-output.txt new-output.txt
    rm -f old-output.txt new-output.txt
done

# Special case for bool arguments

$OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $opt --exclude-system-jobs \
    | sort > old-output.txt
$NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $opt --exclude-system-jobs \
    | sort > new-output.txt
cmp old-output.txt new-output.txt
rm -f old-output.txt new-output.txt
