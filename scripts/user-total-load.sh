#!/bin/bash
#
# This is a template script for reporting per-user resource usage.  The first six parameters need to
# be instantiated for your use case.  Then in the program it is possible you want different awk code
# for the quantity you're working on, for example, SUM_AND_PERCENT does not make a lot of sense with
# memory.
#
# CLUSTER and HOST need to be in sync, clearly.

QUANT=gputime/sec
CLUSTER=ml
HOST=ml8
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=16w
DISCRIMINANT=--some-gpu

SONALYZE=../sonalyze/target/release/sonalyze
REMOTE=http://158.39.48.160:8087

# User total load across a time period, in absolute terms and as a percentage of the total.  This
# makes sense for cpu time and gpu time, at least.

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

#

$SONALYZE jobs \
  --auth-file $AUTH \
  --cluster $CLUSTER \
  --remote $REMOTE \
  -u- \
  --fmt=awk,user,$QUANT \
  --from $TIMESPAN \
  --host $HOST \
  $DISCRIMINANT | \
    awk "$SUM_AND_PERCENT" | \
    sort -k 2nr
