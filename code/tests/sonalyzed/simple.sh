#!/bin/bash

set -e

# Note if `sonalyze daemon` fails on startup the `set -e` will not catch it because the server is
# run in the background.  In this case, $sonalyzed_pid will reference a process that is not there.

rootdir=test-root
testapi=127.0.0.1:4545
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
           -rest-api $testapi \
           -v0 \
           -v1 \
           -insert \
           -upload-auth upload-auth.txt \
           -analysis-auth analysis-auth.txt &
sonalyzed_pid=$!

# Always attempt to shut down the server on exit.  (Not sure if the HUP/INT are necessary or if they
# are subsumed by EXIT.)
trap "kill -HUP $sonalyzed_pid" EXIT ERR SIGHUP SIGINT

# Wait for sonalyzed to come up
sleep 1

# First, try to insert some data and verify that the data have been added as expected

curl --silent --fail-with-body --data-binary @cluster1-samples.json -H 'Content-Type: application/json' -u cluster1.naic.com:hohoho \
     http://$testapi/api/v1/insert/sample > /dev/null

curl --silent --fail-with-body --data-binary @cluster1-sysinfo.json -H 'Content-Type: application/json' -u cluster1.naic.com:hohoho \
     http://$testapi/api/v1/insert/sysinfo > /dev/null

curl --silent --fail-with-body --data-binary @cluster2-samples.json -H 'Content-Type: application/json' -u cluster2.naic.com:hahaha \
     http://$testapi/api/v1/insert/sample > /dev/null

curl --silent --fail-with-body --data-binary @cluster2-sysinfo.json -H 'Content-Type: application/json' -u cluster2.naic.com:hahaha \
     http://$testapi/api/v1/insert/sysinfo > /dev/null

sleep 1

# Note input data tagged say 2023-09-15T00:00:nn+02:00 are normalized to 2023-09-14.  Etc.
#
# Also note that the /api/v1 ingestion path *parses* the input and then serializes it again, so the
# results are not bitwise comparable (unlike before).  This is not ideal but was an inevitable
# consequence of the rewrite.  In any case, we must use jq here to sort the fields before we compare
# the files.

jq -S < cluster1-samples.json > $rootdir/a
jq -S < $rootdir/data/cluster1.naic.com/2026/04/29/0+sample-slurm-monitor.uio.no.json > $rootdir/b
cmp $rootdir/a $rootdir/b

jq -S < cluster1-sysinfo.json > $rootdir/a
jq -S < $rootdir/data/cluster1.naic.com/2026/04/29/0+sysinfo-slurm-monitor.uio.no.json > $rootdir/b
cmp $rootdir/a $rootdir/b

jq -S < cluster2-samples.json > $rootdir/a
jq -S < $rootdir/data/cluster2.naic.com/2026/04/29/0+sample-naic-monitor.uio.no.json > $rootdir/b
cmp $rootdir/a $rootdir/b

jq -S < cluster2-sysinfo.json > $rootdir/a
jq -S < $rootdir/data/cluster2.naic.com/2026/04/29/0+sysinfo-naic-monitor.uio.no.json > $rootdir/b
cmp $rootdir/a $rootdir/b

# Then, try to run queries and verify that the result is what we expect.  API 0 returns a JSON
# string that must be parsed to get the expected output.

output=$(curl --silent --fail-with-body -G -u john:jj \
              "http://127.0.0.1:4545/api/v0/node?from=2026-04-29&cluster=cluster1.naic.com&fmt=default,csv,noheader" \
             | jq -r)
CHECK "node_1" 'slurm-monitor.uio.no,4,47,0,0,"4x1 Intel(R) Xeon(R) Gold 6448Y, 47 GiB"' "$output"

rm -rf $rootdir
