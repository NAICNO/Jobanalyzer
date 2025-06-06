#!/bin/bash

set -e

# No support for `add` in the rust version
if [[ $($SONALYZE version) =~ sonalyze-rs ]]; then
    echo "Skipping sonalyzed tests for sonalyze-rs"
    exit 0
fi

# Note if `sonalyze daemon` fails on startup the `set -e` will not catch it because the server is
# run in the background.  In this case, $sonalyzed_pid will reference a process that is not there.

rootdir=test-root
testport=24680
rm -rf $rootdir

# Set up a test jobanalyzer directory structure

mkdir -p $rootdir $rootdir/cluster-config $rootdir/data/cluster{1,2}.naic.com
cp $SONALYZE $rootdir
cp cluster-aliases.json $rootdir/cluster-config
cp cluster1.naic.com-config.json $rootdir/cluster-config/cluster1.naic.com-config.json
cp cluster2.naic.com-config.json $rootdir/cluster-config/cluster2.naic.com-config.json

# Run the server in the background against that directory

$SONALYZE daemon -v \
           -jobanalyzer-dir $rootdir \
           -port $testport \
           -upload-auth upload-auth.txt \
           -analysis-auth analysis-auth.txt \
           -match-user-and-cluster &
sonalyzed_pid=$!

# Always attempt to shut down the server on exit.  (Not sure if the HUP/INT are necessary or if they
# are subsumed by EXIT.)
trap "kill -HUP $sonalyzed_pid" EXIT ERR SIGHUP SIGINT

# Wait for sonalyzed to come up
sleep 1

# First, try to insert some data and verify that the data have been added as expected

curl --fail-with-body --data-binary @cluster1-samples.csv -H 'Content-Type: text/csv' -u cluster1.naic.com:hohoho \
     http://localhost:$testport/sonar-freecsv?cluster=cluster1.naic.com

curl --fail-with-body --data-binary @cluster1-sysinfo.json -H 'Content-Type: application/json' -u cluster1.naic.com:hohoho \
     http://localhost:$testport/sysinfo?cluster=cluster1.naic.com

curl --fail-with-body --data-binary @cluster2-samples.csv -H 'Content-Type: text/csv' -u cluster2.naic.com:hahaha \
     http://localhost:$testport/sonar-freecsv?cluster=cluster2.naic.com

curl --fail-with-body --data-binary @cluster2-sysinfo.json -H 'Content-Type: application/json' -u cluster2.naic.com:hahaha \
     http://localhost:$testport/sysinfo?cluster=cluster2.naic.com

sleep 1
# Note input data tagged say 2023-09-15T00:00:nn+02:00 are normalized to 2023-09-14.  Etc.
cmp cluster1-samples.csv $rootdir/data/cluster1.naic.com/2023/09/14/c1.cluster1.naic.com.csv
cmp cluster1-sysinfo.json $rootdir/data/cluster1.naic.com/2024/03/11/sysinfo-c1.cluster1.naic.com.json
cmp cluster2-samples.csv $rootdir/data/cluster2.naic.com/2023/09/13/c2.cluster2.naic.com.csv
cmp cluster2-sysinfo.json $rootdir/data/cluster2.naic.com/2024/03/31/sysinfo-c2.cluster2.naic.com.json

# Then, try to run a jobs command and verify that the result is what we expect

output=$(curl --silent --fail-with-body -G -u john:jj \
              "http://localhost:$testport/jobs?cluster=cluster1.naic.com&job=2712710&from=2023-09-01&fmt=noheader,csv,std,cpu,mem")
CHECK "jobs_1" "2712710!,hermanno,0d0h20m,c1.cluster1.naic.com,3,7,14,14" "$output"

rm -rf $rootdir
