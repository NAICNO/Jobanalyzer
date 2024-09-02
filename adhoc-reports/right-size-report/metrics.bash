#!/bin/bash
#
# This should create metrics-10,20,30.png in the current directory, based on Slurm data.

source config.bash

# The order of output fields is assumed by metrics.py.  It does not use all of them but most of the
# rest are used for filtering by the grep below.
FIELDS=rcpu,rmem,JobID,User,Account,End,State,JobName

$SONALYZE_SACCT -from 2024-03-01 -to 2024-05-31 \
                -min-runtime 2h -min-reserved-cores 32 -min-reserved-mem 20 \
                -fmt awk,$FIELDS \
    | grep -v -E 'interactive|OOD|ood|TIMEOUT|CANCELLED' \
    | python metrics.py 10 20 30 \
    | gnuplot


