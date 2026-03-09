#!/usr/bin/env bash
#
# Print sacct-like data synthesized from Sonar data for jobs from non-slurm systems and non-slurm
# jobs on slurm systems.  This is intended to be an example, not a feature-complete application.
#
# -c cluster-name
#   The canonical name of the cluster
#
# -d database-uri
#   The timescaledb URI, but ALWAYS using a postgresql: scheme
#
# -w date
#   Set the from and to dates to the same date, yyyy-mm-dd
#
# -f date
#   Set the from date (inclusive), yyyy-mm-dd
#
# -t date
#   Set the to date (inclusive), yyyy-mm-dd
#
# Time stamps are always UTC.
#
# See below for various important notes.

USAGE="Usage: $0 -c cluster-name -d database-uri [-w date] [-f date] [-t date]"

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
if [[ -z $CLUSTER_NAME || -z $DATABASE_URI ]]; then
    echo "Missing required argument."
    echo $USAGE
    exit 1
fi

# The logic here is that we're basically getting the effect of a `sonar jobs` run with a large time
# window, some multiple of 24h.  The output format will be different from Sonar's, of course.  Job
# state reflects the state at the end of the window.  No jobs are PENDING or CANCELLED - only
# RUNNING or COMPLETED.
#
# The only fields here are the ones that are computable from Sonar data.
#
# Note:
#
# - All time stamps are Unix seconds
#
# - MinCPU is basically pointless if Sonar rolls up jobs, esp MPI jobs.
#
# - Sonalyze currently (2026-03-05) mostly ignores the Epoch field of process samples, so it is
#   possible but very unlikely that what are actually separate jobs on the same node may be confused
#   as parts of the same job.  For this to happen, the two jobs must both run in the time window
#   being queried, on the same host, have the same process group ID, and the same command name.
#
# - A job that starts before the time window or ends after the time window will have partial
#   information here, in particular, a job that started before the time window and ended in the time
#   window will show as COMPLETED but the fields will only reflect the part of the job that is
#   visible in the time window!  To retrieve complete information for such a job, the client MUST
#   step back in time to find the start of the job and then run a query for the job for the entire
#   time window it is running.  Even for one-day time windows, many (most?) jobs will start and end
#   in the time window; so having to iterate need not be bad.  We print the Primordial flag here,
#   which is set if a job was believed to be alive at the beginning of the time window.  There are
#   other flags that may also be helpful.
#
#   It would be possible to enrich Sonalyze with the logic to search back in time for the start of
#   jobs that are found in the initial time window (and for that matter, to search forward for the
#   end, if the end of the window is in the past), removing that burden from the client code.
#
# - The output format "native" is JSON with numbers represented as numbers, not as strings (unlike
#   the normal JSON format, for historical reasons).

${SONALYZE:-sonalyze} jobs \
                      -database-uri ${DATABASE_URI} \
                      -cluster ${CLUSTER_NAME} \
                      -user - \
                      -sacct-from-sonar \
                      -from ${FROM} \
                      -to ${TO} \
                      -fmt native,AveCPU,AveDiskRead,AveDiskWrite,AveRSS,AveVMSize,ElapsedRaw,End,JobID,JobName,MaxRSS,MaxVMSize,MinCPU,NodeList,Primordial,ReqCPUS,ReqGPUS,ReqMem,Start,State,Submit,Time,UserCPU,User

