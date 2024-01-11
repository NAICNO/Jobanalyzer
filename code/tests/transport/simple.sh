set -e

testport=24680
tmpdir=./simple-data
rm -rf $tmpdir
mkdir -p $tmpdir

# If we exit early because of the `set -e` or for other reasons then the infiltrate process will not
# be killed properly and subsequent attempts to run will fail because the port is in use.

# Note if infiltrate fails on startup the `set -e` will not catch it because the server is run in
# the background.  In this case, $infiltrate_pid will reference a process that is not there.

$INFILTRATE -data-path $tmpdir -port $testport -auth-file test-auth.txt &
infiltrate_pid=$!

# Always attempt to shut down the server on exit.  (Not sure if the HUP/INT are necessary or if they
# are subsumed by EXIT.)
trap "kill -HUP $infiltrate_pid" EXIT ERR SIGHUP SIGINT

# Wait for infiltrate to come up
sleep 1

cat simple-data1.csv | \
    $EXFILTRATE \
        -window 0 \
        -cluster test \
        -source sonar/csvnamed \
        -output json \
        -target http://localhost:$testport \
        -auth-file test-auth.txt
cat simple-data2.csv | \
    $EXFILTRATE \
        -window 0 \
        -cluster test \
        -source sonar/csvnamed \
        -output json \
        -target http://localhost:$testport \
        -auth-file test-auth.txt

# Wait for messages to be processed by infiltrate
sleep 1

CHECK transport_simple "$(cat simple-expect.csv)" "$(sort $tmpdir/test/2023/11/15/myhost.csv)"

rm -rf $tmpdir
