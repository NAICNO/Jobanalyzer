#!/usr/bin/env bash

USAGE="Usage: job-data.bash -c cluster-name -d database-uri"

DATABASE_URI=
CLUSTER_NAME=
while getopts c:d:h opt $@; do
    case $opt in
        c) CLUSTER_NAME=$OPTARG ;;
        d) DATABASE_URI=$OPTARG ;;
        h) echo $USAGE; exit 0 ;;
        *) exit 1 ;;
    esac
done
if [[ -z $CLUSTER_NAME || -z $DATABASE_URI ]]; then
    echo "Missing required argument."
    echo $USAGE
    exit 1
fi

# The logic here is that we're basically getting a `sonar jobs` run with a large time window, some
# multiple of 24h.  The output format will be different, of course.  Job state reflects the state at
# the end of the window.  No jobs are PENDING or CANCELLED - only RUNNING or COMPLETED.
#
# TODO
# - "Time" should be timestamp of earliest record in data set

${SONALYZE:-sonalyze} jobs \
                      -database-uri ${DATABASE_URI} \
                      -cluster ${CLUSTER_NAME} \
                      -user - \
                      -sacct-from-sonar \
                      -fmt json,AveCPU,AveDiskRead,AveDiskWrite,AveRSS,AveVMSize,ElapsedRaw,End,JobID,JobName,MaxRSS,MaxVMSize,MinCPU,NodeList,ReqCPUS,ReqGPUS,ReqMem,Start,State,Submit,Time,UserCPU,User,Version

