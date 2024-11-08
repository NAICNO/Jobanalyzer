#!/usr/bin/bash
#
# hwmon - monitor cluster status across all clusters and alert about clusters that seem to be down
# based on the number of nodes reporting
#
# This uses `sonalyze uptime` which runs off the config file, which is not ideal at present since
# that file tends to be a little stale.

MAILTO=${MAILTO:-larstha@uio.no}

SONALYZE=~/go/bin/sonalyze
REMOTE=https://naic-monitor.uio.no
AUTH=~/.ssh/sonalyzed-auth.netrc

for cluster in mlx fox saga fram betzy; do
    # The problem here is that we have both up and down records and even if the host is up now we'll
    # see a down record from it from earlier in the day (say).  So what we want is to compare `end`
    # times for the same host, and if the latest up end time is later than the latest down end time
    # for the host then it is up, otherwise down.
    $SONALYZE uptime \
              -remote "$REMOTE" -auth-file "$AUTH" \
              -cluster "$cluster" -interval 10 -fmt awk,noheader,device,host,state,end | \
        awk -v "cluster=$cluster" '
$1 == "host" {
  hosts[$2] = 1
  if ($3 == "up") {
    if (latest_up[$2] < $3) {
      latest_up[$2] = $3
      ups++
    }
  } else {
    if (latest_down[$2] < $3) {
      latest_down[$2] = $3
      downs++
    }
  }
}
$1 == "gpu" {}
END {
  if (ups >= downs) {
    print cluster " appears to be down: " ups "/" (ups+downs) " are up"
  }
  for (host in hosts) {
    if (latest_up[host] >= latest_down[host]) {
      #print cluster " " host " is up"
    } else {
      print cluster " " host " is down"
    }
  }
}'
done
