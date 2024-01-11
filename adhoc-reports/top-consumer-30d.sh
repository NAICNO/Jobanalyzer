#!/bin/bash
#
# Standard report (prototype): "top consumers last n days" aka "system load by user"
#
# Currently ML cluster only, but easy to extend once the data are on the analysis host.
# Currently shows entries sorted by descending CPU time or GPU time (in seconds).
#
# The AUTH file must have username:password on one line; the identity must be known to the server.
#
# (Source: This is the same as worst-violators.sh except that this one does not filter by --no-gpu
# or --some-gpu, it looks at all jobs.)

CLUSTER=ml
HOST="ml[1-9]"
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=30d
SORTBY=2      # "2" for CPU time, "4" for GPU time
HOWMANY=25

# Standard configuration
SONALYZE=../code/sonalyze/target/release/sonalyze
REMOTE=http://naic-monitor.uio.no:8087

FIELDS=user,cputime/sec,gputime/sec
SUM_AND_PERCENT='
{
  user = $1
  cpusec = $2
  gpusec = $3
  cputime[user] += cpusec
  cpusum += cpusec
  gputime[user] += gpusec
  gpusum += gpusec
}
END {
  for (user in cputime) {
    printf("%s %d %3.1f %d %3.1f\n",
           user,
           cputime[user],
           cputime[user] * 100 / cpusum,
           gputime[user],
           gputime[user] * 100 / gpusum)
  }
}'

$SONALYZE jobs \
  --remote $REMOTE \
  --cluster $CLUSTER \
  --auth-file $AUTH \
  --user - \
  --fmt=awk,$FIELDS \
  --from $TIMESPAN \
  --host "$HOST" \
    | awk "$SUM_AND_PERCENT" \
    | sort -k ${SORTBY}nr \
    | head -n $HOWMANY
