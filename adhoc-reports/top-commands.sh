#!/bin/bash

CLUSTER=saga
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=1w

SONALYZE=${SONALYZE:-../code/sonalyze/sonalyze}
REMOTE=https://naic-monitor.uio.no

$SONALYZE jobs \
  --auth-file $AUTH \
  --cluster $CLUSTER \
  --remote $REMOTE \
  --user - \
  --fmt=awk,cputime/sec,cmd \
  --from $TIMESPAN | awk '
{
  procs[$2] += $1
  sum += $1
}
END {
  for (j in procs) {
    if (procs[j] != 0) {
      print j, procs[j], procs[j] * 100 / sum
    }
  }
}' | sort -nrk2 | head -n 25

