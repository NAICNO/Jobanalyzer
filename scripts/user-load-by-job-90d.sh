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

# Configure this only if you're running from some other directory or you've not compiled the
# binaries yourself and you're using somebody else's.  To build binaries, run build.sh in the parent
# directory.
SONALYZE=../sonalyze/target/release/sonalyze

# Generally don't mess with these.
QUANT=cputime/sec,gputime/sec,cmd
REMOTE=http://158.39.48.160:8087

SUM_AND_PERCENT='
{
  cputime[$3] += $1
  gputime[$3] += $2
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
  --auth-file $AUTH \
  --cluster $CLUSTER \
  --remote $REMOTE \
  --user $USER \
  --fmt=awk,$QUANT \
  --from $TIMESPAN \
  --host "$HOST" \
  $DISCRIMINANT | \
    awk "$SUM_AND_PERCENT" | \
    sort -k ${SORTBY}nr | head -n $HOWMANY
