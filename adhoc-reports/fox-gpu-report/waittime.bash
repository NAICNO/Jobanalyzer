#!/bin/bash
#
# Note, this is meant to be run on naic-monitor.uio.no against a local data store.

from=$1
from=${from:-2024-01-01}
for host in gpu-1 gpu-2 gpu-7 gpu-8; do
    echo $host
    ./sonalyze sacct \
               -data-dir ~/fox-experiment/data2 \
               -from $from \
               -all \
               -fmt awk,Submit,Start,Wait,NodeList \
        | grep -E "$host\$" \
        | ./waittime.py
done
