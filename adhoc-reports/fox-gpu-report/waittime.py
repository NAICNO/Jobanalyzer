#!/usr/bin/python
#
# Input format:
#  (ignored) (ignored) WaitTimeSec (ignored) ...
#
# Output format (tab-separated)
#  total-jobs 0h 1h 2h 3h 4-11h 12-23h 1-6d > 6d

import sys

def main():
    # Read data and bucket by hours waited
    buckets = {}
    count=0
    for l in sys.stdin:
        fs = l.split()
        w = int(fs[2])
        b = w % 3600
        count += 1
        if not b in buckets:
            buckets[b] = 1
        else:
            buckets[b] = buckets[b] + 1

    waiting0=0
    waiting1=0
    waiting2=0
    waiting3=0
    waiting4=0
    waiting12=0
    waiting_days=0
    waiting_weeks=0
    if 0 in buckets:
        waiting0=buckets[0]
    if 1 in buckets:
        waiting1=buckets[1]
    if 2 in buckets:
        waiting2=buckets[2]
    if 3 in buckets:
        waiting3=buckets[3]
    for i in range(4,12):
        if i in buckets:
            waiting4 += buckets[i]
    for i in range(12,23):
        if i in buckets:
            waiting12 += buckets[i]
    for k in buckets:
        if k > 24 and k < 24*7:
            waiting_days += buckets[k]
        elif k >= 24.7:
            waiting_weeks += buckets[k]

    print(str(count) + "\t" +
          str(waiting0) + "\t" +
          str(waiting1) + "\t" +
          str(waiting2) + "\t" +
          str(waiting3) + "\t" +
          str(waiting4) + "\t" +
          str(waiting12) + "\t" +
          str(waiting_days) + "\t" +
          str(waiting_weeks))

main()
