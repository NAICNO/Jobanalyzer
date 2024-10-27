#!/bin/bash
#
# Probably want a "tag" that can be used to identify the run as to input data and machine.

set -e

# This stress-tests parsing, formatting, and stream synthesis.
for fmt in "--fmt=fixed,all" "--fmt=csv,all" "--fmt=csvnamed,all" "--fmt=awk,all" "--fmt=json,all"; do
    for opt in "" --clean --merge-by-host-and-job --merge-by-job; do
        /usr/bin/time -f "parse,$OLD_NAME,\"$fmt\",$opt,$(hostname),%E,%M" \
                      $OLD_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $opt $fmt > /dev/null
        /usr/bin/time -f "parse,$NEW_NAME,\"$fmt\",$opt,$(hostname),%E,%M" \
                      $NEW_SONALYZE parse --data-path "$DATA_PATH" --from "$FROM" --to "$TO" $opt $fmt > /dev/null
    done
done

# For the rest, just make sure they don't drool

/usr/bin/time -f "jobs,$OLD_NAME,,,$(hostname),%E,%M" \
              $OLD_SONALYZE jobs -u - --data-path "$DATA_PATH" --from "$FROM" --to "$TO" > /dev/null
/usr/bin/time -f "jobs,$NEW_NAME,,,$(hostname),%E,%M" \
              $NEW_SONALYZE jobs -u - --data-path "$DATA_PATH" --from "$FROM" --to "$TO" > /dev/null

/usr/bin/time -f "load,$OLD_NAME,,,$(hostname),%E,%M" \
              $OLD_SONALYZE load --data-path "$DATA_PATH" --from "$FROM" --to "$TO" > /dev/null
/usr/bin/time -f "load,$NEW_NAME,,,$(hostname),%E,%M" \
              $NEW_SONALYZE load --data-path "$DATA_PATH" --from "$FROM" --to "$TO" > /dev/null

/usr/bin/time -f "uptime,$OLD_NAME,,,$(hostname),%E,%M" \
              $OLD_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 5 > /dev/null
/usr/bin/time -f "uptime,$NEW_NAME,,,$(hostname),%E,%M" \
              $NEW_SONALYZE uptime --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --interval 5 > /dev/null

/usr/bin/time -f "profile,$OLD_NAME,,,$(hostname),%E,%M" \
              $OLD_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --job "$JOB" > /dev/null
/usr/bin/time -f "profile,$NEW_NAME,,,$(hostname),%E,%M" \
              $NEW_SONALYZE profile --data-path "$DATA_PATH" --from "$FROM" --to "$TO" --job "$JOB" > /dev/null
