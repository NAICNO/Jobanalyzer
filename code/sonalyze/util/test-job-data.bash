#!/usr/bin/env bash
#
# Same arguments as to job-data.bash, q.v.

set -e
set -o pipefail
DATABASE_URI=
CLUSTER_NAME=
FROM=$(date +%F)
TO=$FROM
while getopts c:d:f:t:w:h opt $@; do
    case $opt in
        c) CLUSTER_NAME=$OPTARG ;;
        d) DATABASE_URI=$OPTARG ;;
        f) FROM=$OPTARG ;;
        t) TO=$OPTARG ;;
        w) FROM=$OPTARG ; TO=$OPTARG ;;
        h) echo $USAGE; exit 0 ;;
        *) echo $USAGE; exit 1 ;;
    esac
done

./job-data.bash -c "$CLUSTER_NAME" -d "$DATABASE_URI" -f "$FROM" -t "$TO" > test-data.json

# This should be enough to read everything and check the syntax
jq '.[] | .AveCPU' test-data.json | wc -l
rm -f test-data.json
echo "OK"
