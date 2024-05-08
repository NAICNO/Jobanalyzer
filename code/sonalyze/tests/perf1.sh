#!/bin/bash
#
# Probably want a "tag" that can be used to identify the run as to input data and machine.
#
# Usage:
#  perf1.sh data-dir from to job

GO_SONALYZE=${GO_SONALYZE:-../sonalyze}
RUST_SONALYZE=${RUST_SONALYZE:-../../attic/sonalyze/target/release/sonalyze}

set -e

# This stress-tests parsing, formatting, and stream synthesis.
for fmt in "--fmt=fixed,all" "--fmt=csv,all" "--fmt=csvnamed,all" "--fmt=awk,all" "--fmt=json,all"; do
    for opt in "" --clean --merge-by-host-and-job --merge-by-job; do
        /usr/bin/time -f "parse,go,\"$fmt\",$opt,$(hostname),%E,%M" \
                      $GO_SONALYZE parse --data-path "$1" --from "$2" --to "$3" $opt $fmt > /dev/null
        /usr/bin/time -f "parse,rust,\"$fmt\",$opt,$(hostname),%E,%M" \
                      $RUST_SONALYZE parse --data-path "$1" --from "$2" --to "$3" $opt $fmt > /dev/null
    done
done

# For the rest, just make sure they don't drool

/usr/bin/time -f "jobs,go,,,$(hostname),%E,%M" \
              $GO_SONALYZE jobs -u - --data-path "$1" --from "$2" --to "$3" > /dev/null
/usr/bin/time -f "jobs,rust,,,$(hostname),%E,%M" \
              $RUST_SONALYZE jobs -u - --data-path "$1" --from "$2" --to "$3" > /dev/null

/usr/bin/time -f "load,go,,,$(hostname),%E,%M" \
              $GO_SONALYZE load --data-path "$1" --from "$2" --to "$3" > /dev/null
/usr/bin/time -f "load,rust,,,$(hostname),%E,%M" \
              $RUST_SONALYZE load --data-path "$1" --from "$2" --to "$3" > /dev/null

/usr/bin/time -f "uptime,go,,,$(hostname),%E,%M" \
              $GO_SONALYZE uptime --data-path "$1" --from "$2" --to "$3" --interval 5 > /dev/null
/usr/bin/time -f "uptime,rust,,,$(hostname),%E,%M" \
              $RUST_SONALYZE uptime --data-path "$1" --from "$2" --to "$3" --interval 5 > /dev/null

/usr/bin/time -f "profile,go,,,$(hostname),%E,%M" \
              $GO_SONALYZE profile --data-path "$1" --from "$2" --to "$3" --job "$4" > /dev/null
/usr/bin/time -f "profile,rust,,,$(hostname),%E,%M" \
              $RUST_SONALYZE profile --data-path "$1" --from "$2" --to "$3" --job "$4" > /dev/null
