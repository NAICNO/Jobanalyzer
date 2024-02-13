#!/bin/bash
#
# Standard report (prototype): "rank user jobs by load"
#
# Currently ML cluster only, but easy to extend once the data are on the analysis host.
# Currently shows entries sorted by descending CPU time or GPU time (in seconds).
#
# The AUTH file must have username:password on one line; user must be known to the server.

# Configure these as necessary
CLUSTER=ml
HOST="ml[1-9]"
USER=pubuduss
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=90d
DISCRIMINANT= # Use --no-gpu or --some-gpu if you like
SORTBY=2      # "1" for CPU time, "2" for GPU time
HOWMANY=25

# Standard configuration
SONALYZE=../code/sonalyze/target/release/sonalyze
REMOTE=https://naic-monitor.uio.no

FIELDS=cputime/sec,gputime/sec,cmd
SUM_AND_PERCENT='
{
  cpusec=$1
  gpusec=$2
  cmd=$3
  cputime[cmd] += cpusec
  gputime[cmd] += gpusec
}
END {
  for (cmd in cputime) {
    printf("%d %d %s\n",
           cputime[cmd],
           gputime[cmd],
           cmd)
  }
}'

$SONALYZE jobs \
  --remote $REMOTE \
  --cluster $CLUSTER \
  --auth-file $AUTH \
  --user $USER \
  --fmt=awk,$FIELDS \
  --from $TIMESPAN \
  --host "$HOST" \
  $DISCRIMINANT \
    | awk "$SUM_AND_PERCENT" \
    | sort -k ${SORTBY}nr \
    | head -n $HOWMANY
