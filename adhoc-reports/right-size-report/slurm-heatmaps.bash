#!/bin/bash
#
# Before running this, go to ../../code and run build.sh
#
# This is currently meant to run on naic-report with direct access to the data stores (multiple
# stores, one for production and one for experimental slurm data).  But it should be easy enough to
# change this to remote access to a single store later.

source config.bash
FIELDS=rcpu,rmem,JobName,State

# "All" CPU/RAM data in 10x10 map
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -all -fmt awk,$FIELDS | $HEATMAP -n 10 > all-map.txt

# "Serious" data
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 -fmt awk,$FIELDS | $HEATMAP -n 10 > serious-map.txt

# "Serious" data minus interactive
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 -fmt awk,$FIELDS | grep -v -E 'interactive|OOD|ood' | $HEATMAP -n 10 > serious-noninteractive-map.txt

# Serious noninteractive TIMEOUT jobs
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 -fmt awk,$FIELDS | grep -v -E 'interactive|OOD|ood' | grep TIMEOUT | $HEATMAP -n 10 > serious-timeout-map.txt

# "Serious" data minus interactive or TIMEOUT
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -min-runtime 2h -min-reserved-mem 20 -min-reserved-cores 32 -fmt awk,$FIELDS | grep -v -E 'interactive|OOD|ood|TIMEOUT' | $HEATMAP -n 10 > serious-noninteractive-nontimeout-map.txt

# "Large" data minus interactive or TIMEOUT
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -min-runtime 1d -min-reserved-mem 20 -min-reserved-cores 32 -fmt awk,$FIELDS | grep -v -E 'interactive|OOD|ood|TIMEOUT' | $HEATMAP -n 10 > large-noninteractive-nontimeout-map.txt

# "Large" data minus interactive or TIMEOUT or CANCELLED
$SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY -min-runtime 1d -min-reserved-mem 20 -min-reserved-cores 32 -fmt awk,$FIELDS | grep -v -E 'interactive|OOD|ood|TIMEOUT|CANCELLED' | $HEATMAP -n 10 > large-noninteractive-nontimeout-noncancelled-map.txt
