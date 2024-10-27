#!/bin/bash
#
# Test that a plain run of `sonalyze profile -fmt=<fmt>,<field>` produces bit-identical output between
# versions.  See test-generic.sh for info about env vars.

set -e
for format in csv html fixed; do
    echo "  $format"
    for field in cpu mem gpu res gpumem; do
        echo "    $field"
        for max in 0 50 200; do
            echo "      $max"
            # Go and Rust differ for 0: for Go it means "no argument present" and for Rust it means
            # that the argument is actually 0.  Work around this.
            maxarg=""
            if [[ $max != 0 ]]; then
                maxarg="--max $max"
            fi
            $OLD_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" $maxarg --fmt $format,$field > old-output.txt
            $NEW_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" -j "$JOB" $maxarg --fmt $format,$field > new-output.txt
            cmp old-output.txt new-output.txt
            rm -f old-output.txt new-output.txt
        done
    done
done


