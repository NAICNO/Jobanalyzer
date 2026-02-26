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

${SONALYZE:-sonalyze} jobs \
                      -database-uri ${DATABASE_URI} \
                      -cluster ${CLUSTER_NAME} \
                      -user - \
                      -sacct-from-sonar \
                      -fmt json,Job,Running,Completed,User,Start,Hosts,End,ResidentMemAvgGB,MemAvgGB,Duration,ResidentMemPeakGB,MemPeakGB

