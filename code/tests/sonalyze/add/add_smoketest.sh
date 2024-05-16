#!/bin/bash

set -e

# No support for `add` in the rust version
if [[ $($SONALYZE version) =~ sonalyze-rs ]]; then
    echo "Skipping sonalyze add tests for sonalyze-rs"
    exit 0
fi

rm -rf data-dir
mkdir data-dir

# Make it interesting by adding things in batches

k1=$(wc -l smoketest-sample-data1.csv | awk '{ print $1 }')
k2=$(wc -l smoketest-sample-data2.csv | awk '{ print $1 }')

head -n $((k1/2)) smoketest-sample-data1.csv | $SONALYZE add -v -sample -data-dir ./data-dir
head -n $((k2/2)) smoketest-sample-data2.csv | $SONALYZE add -v -sample -data-dir ./data-dir
tail -n $((k1-k1/2)) smoketest-sample-data1.csv | $SONALYZE add -v -sample -data-dir ./data-dir
tail -n $((k2-k2/2)) smoketest-sample-data2.csv | $SONALYZE add -v -sample -data-dir ./data-dir

$SONALYZE add -v -sysinfo -data-dir ./data-dir < smoketest-sysinfo-data1.json
$SONALYZE add -v -sysinfo -data-dir ./data-dir < smoketest-sysinfo-data2.json

for d in data-dir/2023/09/{13,15} data-dir/2024/03/12 data-dir/2024/04/01; do
    if [[ ! -d $d ]]; then
        echo "Failed to find directory $d"
        exit 1
    fi
done

cmp data-dir/2023/09/13/ml3.hpc.uio.no.csv smoketest-sample-data2.csv
cmp data-dir/2023/09/15/ml6.hpc.uio.no.csv smoketest-sample-data1.csv
cmp data-dir/2024/03/12/sysinfo-ml1.hpc.uio.no.json smoketest-sysinfo-data1.json
cmp data-dir/2024/04/01/sysinfo-ml9.hpc.uio.no.json smoketest-sysinfo-data2.json

rm -rf data-dir
