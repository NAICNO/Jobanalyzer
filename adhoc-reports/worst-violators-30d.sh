#!/bin/bash
#
# Standard report (prototype): "worst violators of ML policy last 30 days"
#
# Users are listed in descending order of CPU time used by their jobs that did not use GPU at all,
# and the percentage of their CPU time of the CPU time of all such jobs.
#
# The AUTH file must have username:password on one line; the identity must be known to the server.

CLUSTER=ml
HOST="ml[1-9]"
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=30d
HOWMANY=25

# Standard configuration
SONALYZE=../code/sonalyze/target/release/sonalyze
REMOTE=https://naic-monitor.uio.no

FIELDS=user,cputime/sec
SUM_AND_PERCENT='
{
  user=$1
  cpusec=$2
  cputime[user] += cpusec
  cpusum += cpusec
}
END {
  for (user in cputime) {
    if (cputime[user] != 0) {
      print user, cputime[user], cputime[user] * 100 / cpusum
    }
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
  --no-gpu \
    | awk "$SUM_AND_PERCENT" \
    | sort -k 2nr \
    | head -n $HOWMANY

