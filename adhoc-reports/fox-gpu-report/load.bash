#!/bin/bash

from=$1
from=${from:-2024-01-01}

to=$2
to=${to:-2024-08-19}

hosts="gpu-1 gpu-2 gpu-7 gpu-8"

for host in $hosts; do
    echo $host
    ./sonalyze load \
             -remote https://naic-monitor.uio.no \
             -cluster fox \
             -auth-file ~/.ssh/sonalyzed-auth.txt \
             -host "$host" \
             -daily \
             -from "$from" \
             -to "$to" \
             -user - \
             -fmt awk,date,time,cpu,res,gpu,gpumem,rgpu,rgpumem \
        | ./load.py
done
