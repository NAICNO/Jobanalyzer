#!/bin/bash
#
# Standard report (prototype): "top consumers last n days" aka "system load by user"
#
# Currently ML cluster only, but easy to extend once the data are on the analysis host.
# Currently shows entries sorted by descending CPU time or GPU time (in seconds).
#
# The AUTH file must have username:password on one line; user must be known to the server.

# Configure these as necessary
CLUSTER=ml
HOST="ml[1-9]"
AUTH=~/.ssh/sonalyzed-auth.txt
TIMESPAN=30d
DISCRIMINANT= # Use --no-gpu or --some-gpu if you like
SORTBY=2      # "2" for CPU time, "4" for GPU time
HOWMANY=25

# Configure this only if you're running from some other directory or you've not compiled the
# binaries yourself and you're using somebody else's.  To build binaries, run build.sh in the parent
# directory.
SONALYZE=../sonalyze/target/release/sonalyze

# Generally don't mess with these.
QUANT=cputime/sec,gputime/sec
REMOTE=http://158.39.48.160:8087

SUM_AND_PERCENT='
{
  cputime[$1] += $2
  cpusum += $2
  gputime[$1] += $3
  gpusum += $3
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
  --auth-file $AUTH \
  --cluster $CLUSTER \
  --remote $REMOTE \
  -u- \
  --fmt=awk,user,$QUANT \
  --from $TIMESPAN \
  --host "$HOST" \
  $DISCRIMINANT | \
    awk "$SUM_AND_PERCENT" | \
    sort -k ${SORTBY}nr | head -n $HOWMANY
