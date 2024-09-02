#!/usr/bin/python3
#
# The input is the output from `sonalyze load` with awk formatting and
# these fields:
#
#  (ignored) (ignored) (ignored) (ignored) (ignored) (ignored) rgpu (ignored)...
#
# eg this command:
#
# sonalyze load \
#   -remote https://naic-monitor.uio.no \
#   -cluster fox \
#   -auth-file ~/.ssh/sonalyzed-auth.txt \
#   -host 'gpu-8' \
#   -daily \
#   -from 2024-01-01 \
#   -fmt awk,date,time,cpu,res,gpu,gpumem,rgpu,rgpumem
#
# The output is one line of text, tab-separated
#
#  average-utilization days-below-40 days-above-80

import sys

n=0
sum=0
n_above_80=0
n_below_40=0
for l in sys.stdin:
    fs = l.split()
    k = int(fs[6])
    n += 1
    sum += k
    if k < 40:
        n_below_40 += 1
    elif k > 80:
        n_above_80 += 1
print(str(sum/n) + "\t" + str(n_below_40) + "\t" + str(n_above_80))
