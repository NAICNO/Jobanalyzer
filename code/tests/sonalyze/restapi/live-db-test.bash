#!/usr/bin/env bash
#
# The idea here is that we run the sonalyze rest API against live data stores and then use that API.
# It's just a smoketest: can we run with default args and not crash?
#
# Then we should redo with various time parameters.

set -e
set -o pipefail

USAGE="Usage: $0 -c cluster-name -d database-uri -j jobanalyzer-dir -D data-directory [-h] [-i interface] [-n node] [-v]
-c is required
Exactly one of -d -j -D is required"

DATABASE_URI=
JOBANALYZER_DIR=
DATA_DIR=
INTERFACE=127.0.0.1:8888
CLUSTER=zappa.yes.no
NODE=
DEFAULT_SONALYZE=../../../sonalyze/sonalyze
verbose=

while getopts c:d:hi:j:n:vD: opt $@; do
    case $opt in
        c) CLUSTER=$OPTARG ;;
        d) DATABASE_URI=$OPTARG ;;
        h) echo $USAGE; exit 0 ;;
        i) INTERFACE=$OPTARG ;;
        j) JOBANALYZER_DIR=$OPTARG ;;
        n) NODE=$OPTARG ;;
        D) DATA_DIR=$OPTARG ;;
        v) verbose=-v ;;
        *) echo $USAGE; exit 1 ;;
    esac
done

# I crashed naic-monitor by sucking up all the memory with an infinite loop the first time, let's
# not do that again.  This is a less than perfect workaround but it should allow us to kill the test
# manually.
export GOMEMLIMIT=1GiB
export SONALYZE_REST_VERBOSE=1

# First start it up
if [[ -n $DATABASE_URI ]]; then
    ${SONALYZE:-$DEFAULT_SONALYZE} daemon -port 24687 -database-uri "$DATABASE_URI" -rest-api "$INTERFACE" $verbose &
    pid=$!
elif [[ -n $JOBANALYZER_DIR ]]; then
    ${SONALYZE:-../../sonalyze/sonalyze} daemon -port 24687 -jobanalyzer-dir "$JOBANALYZER_DIR" -rest-api "$INTERFACE" $verbose &
    pid=$!
elif [[ -n $DATA_DIR ]]; then
    # TODO: Force interpretation of the cluster name
    ${SONALYZE:-../../sonalyze/sonalyze} daemon -port 24687 -data-dir "$DATA_DIR" -rest-api "$INTERFACE" $verbose &
    pid=$!
else
    echo "Missing data source"
    exit 1
fi
echo "Running daemon: $pid"
sleep 5

# Then run things against the REST API.  For some errors (500, 400, 404) this needs to error out,
# but ok to just do visual inspection for now.

smoketest() {
    if [[ -n $verbose ]]; then
        echo $1
    fi
    # Run JQ to ensure output is well-formed
    #curl -s -G "$1" | wc
    curl -s -o /dev/null -w "%{http_code}" $1
    echo
}

# smoketest http://$INTERFACE/api/v2/cluster
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/error-messages
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/cpu/timeseries
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/diskstats/timeseries
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/gpu/timeseries
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/info
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/last-probe-timestamp
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/memory/timeseries
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/process/gpu/util
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/processes
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/processes/gpu
# smoketest http://$INTERFACE/api/v2/cluster/$CLUSTER/processes/timeseries

# 2026-03-17 10:09:40 UTC
START=1773742180

# 2026-03-18 10:09:40 UTC
END=1773828580

# echo "Test /cluster"
# curl -s http://$INTERFACE/api/v2/cluster | jq '.[] | .cluster'

# echo "Test /nodes/memory/timeseries"
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/memory/timeseries?start_time_in_s=$START&end_time_in_s=$END&resolution_in_s=3600"

# echo "Test /nodes/cpu/timeseries"
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/cpu/timeseries?start_time_in_s=$START&end_time_in_s=$END&resolution_in_s=3600"

echo "Test /nodes/gpu/timeseries"
# END=$((START+3600))
curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/gpu/timeseries?start_time_in_s=$START&end_time_in_s=$END&resolution_in_s=3600"

# echo "Test /nodes/info"
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/info"

# echo "Test /nodes/last-probe-timestamp"
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/last-probe-timestamp"

# echo "Test /nodes/process/gpu/util"
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/process/gpu/util?reference_time_in_s=$START&window_in_s=3600"

# echo "Test /processes/timeseries"
# if [[ -z $NODE ]]; then
#     echo "For this you want to ask for one node"
#     exit 1
# fi
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/processes/timeseries?start_time_in_s=$START&end_time_in_s=$END&resolution_in_s=3600&nodename=$NODE"

# Be sure to use fox data here, it has newer sonar
# echo "Test /nodes/diskstats/timeseries"
# if [[ -z $NODE ]]; then
#     echo "For this you want to ask for one node"
#     exit 1
# fi
# curl "http://$INTERFACE/api/v2/cluster/$CLUSTER/nodes/diskstats/timeseries?start_time_in_s=$START&end_time_in_s=$END&resolution_in_s=3600&nodename=$NODE"

echo "Done.  Killing server"
kill $pid
