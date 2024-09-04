#!/bin/bash
#
# See README.md for setup instructions.
#
# You can change SOURCE to "-data-dir ... -config-file ..." for local runs.

SOURCE="-remote https://naic-monitor.uio.no -cluster fox -auth-file $HOME/.ssh/sonalyzed-auth.txt"
SONALYZE=~/go/bin/sonalyze
HEATMAP=~/go/bin/heatmap
FIRSTDAY=2024-03-01
LASTDAY=2024-05-31
FIELDS=rcpu,rmem,JobName,State

# "All" CPU/RAM data in 10x10 map
$SONALYZE sacct $SOURCE \
           -from $FIRSTDAY -to $LASTDAY -all -fmt awk,$FIELDS \
    | $HEATMAP -n 10 > all-map.txt

# "Serious" data
$SONALYZE sacct $SOURCE \
           -from $FIRSTDAY -to $LASTDAY \
           -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 \
           -fmt awk,$FIELDS \
    | $HEATMAP -n 10 > lm-serious-map.txt

# "Serious" data minus interactive
$SONALYZE sacct $SOURCE \
           -from $FIRSTDAY -to $LASTDAY \
           -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 \
           -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood' \
    | $HEATMAP -n 10 > lm-serious-noninteractive-map.txt

# Serious noninteractive TIMEOUT jobs
$SONALYZE sacct $SOURCE -from $FIRSTDAY -to $LASTDAY \
           -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 \
           -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood' \
    | grep TIMEOUT \
    | $HEATMAP -n 10 > lm-serious-timeout-map.txt

# "Serious" data minus interactive or TIMEOUT
$SONALYZE sacct $SOURCE -from $FIRSTDAY -to $LASTDAY \
           -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 \
           -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood|TIMEOUT' \
    | $HEATMAP -n 10 > lm-serious-noninteractive-nontimeout-map.txt

# "Large" data minus interactive or TIMEOUT
$SONALYZE sacct $SOURCE -from $FIRSTDAY -to $LASTDAY \
           -min-runtime 1d -min-reserved-mem 20 -min-reserved-cores 32 \
           -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood|TIMEOUT' \
    | $HEATMAP -n 10 > lm-large-noninteractive-nontimeout-map.txt

# "Large" data minus interactive or TIMEOUT or CANCELLED
$SONALYZE sacct $SOURCE -from $FIRSTDAY -to $LASTDAY \
           -min-runtime 1d -min-reserved-mem 20 -min-reserved-cores 32 \
           -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood|TIMEOUT|CANCELLED' \
    | $HEATMAP -n 10 > lm-large-noninteractive-nontimeout-noncancelled-map.txt
