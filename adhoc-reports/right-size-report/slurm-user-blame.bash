#!/bin/bash

source config.bash
FIELDS=rcpu,rmem,JobName,State,User

# "Serious" data minus interactive or TIMEOUT or CANCELLED
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY \
                -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 \
                -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood|TIMEOUT|CANCELLED' \
    | awk '
{ if ($1 <= 25 && $2 <= 25) {
    blame[$5]++
  }
}
END {
  for (i in blame) {
    print i "\t" blame[i]
  }
}' | sort -nk2 -r | head -n 10 > serious-user-blame.txt
