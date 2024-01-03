#!/bin/bash
#
# See user-total-load.sh for the origin of this script.
#
# List users in descending order by how much CPU time their jobs have used without using any GPU
# time, over the last month by default.  The percentage is a percentage of all the jobs without GPU
# usage, not all jobs - so use that with care.

QUANT=cputime/sec
CLUSTER=ml
HOST="ml[1-9]"
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=4w
DISCRIMINANT=--no-gpu

SONALYZE=../sonalyze/target/release/sonalyze
REMOTE=http://158.39.48.160:8087

SUM_AND_PERCENT='
{
  procs[$1] += $2
  sum += $2
}
END {
  for (j in procs) {
    if (procs[j] != 0) {
      print j, procs[j], procs[j] * 100 / sum
    }
  }
}'

$SONALYZE jobs \
  --auth-file $AUTH \
  --cluster $CLUSTER \
  --remote $REMOTE \
  -u- \
  --fmt=awk,user,$QUANT \
  --from $TIMESPAN \
  --host "$HOST" \
  $DISCRIMINANT | \
    awk "$SUM_AND_PERCENT" | \
    sort -k 2nr
