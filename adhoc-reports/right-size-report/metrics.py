#!/usr/bin/python
#
# Normally this is invoked by metrics.bash, which will provide the correct input and handle the
# output.
#
# Usage:
#
#   metrics.py cutoff ... < input
#
# where the cutoffs are percentages and the output will be a plot of values for all input records
# where both of the first two fields are less than or equal the cutoff, one line for each cutoff.
#
# The input format should be (space-separated fields):
#
#  a b (ignored) (ignored) (ignored) end (ignored) ...
#
# where `end` shall be an ISO formatted timestamp where a T delineates the YYYY-MM-DD format day
# from the time.
#
# The output is a gnuplot program that will create a plot with dates on the x axis and percentage
# values on the y axis.  At any date, the y value is the 7-day running average of y values ending on
# that date, for the given cutoff.  The output is PNG format and is stored in the file
# "metrics-<cutoff>.png" where the <cutoff> is "c1,c2" etc if there are multiple cutoffs.

import sys

# field offsets in the line
RCPU=0
RMEM=1
JOB=2
USER=3
ACCT=4
END=5

cutoffs=[]

def main():
    lines=[]
    for l in sys.stdin:
        lines.append(l.split())

    # Generate a gnuplot program

    lines.sort(key=lambda x: x[END])

    days=[]
    prev=""
    for r in lines:
        day=r[END].split('T')[0]
        if day != prev:
            days.append(day)
            prev=day

    tics=""
    for l in range(len(days)):
        # Not every label
        if l % 7 != 0:
            continue
        if l > 0:
            tics += ", "
        tics += "\"" + days[l] + "\" " + str(l)

    n=str(cutoffs[0])
    for c in cutoffs[1:]:
        n += "," + str(c)
    
    print("set terminal png size 1280,960")
    print("set output \"metrics-" + n + ".png\"")
    print("set xtics rotate (" + tics + ")")
    print("set yrange [0:100]")

    plotters=""
    for cutoff in cutoffs:
        count=[]
        tot=[]
        prev=""
        for r in lines:
            day=r[END].split('T')[0]
            if day != prev:
                count.append(0)
                tot.append(0)
                prev=day

            tot[len(tot)-1] += 1
            c=int(r[RCPU])
            m=int(r[RMEM])
            if c <= cutoff and m <= cutoff:
                count[len(count)-1]+=1

        print("$Data" + str(cutoff) + " << EOD")
        for i in range(len(tot)):
            sum=0
            n=0
            for k in range(max(0, i-6),i+1):
                sum += count[k]/tot[k]
                n += 1
            avg=sum/n
            print(i, int(avg*100))
        print("EOD")

        if plotters != "":
            plotters += ", "
        plotters += "$Data" + str(cutoff) + " with lines lw 2 title \"" + str(cutoff) + "%\""

    print("plot " + plotters)

for a in sys.argv[1:]:
    try:
        cutoff=int(a)
        cutoffs.append(cutoff)
    except:
        print("Bad cutoff value")
        sys.exit(1)

if len(cutoffs) == 0:
    print("Need at least one cutoff")
    sys.exit(1)

main()

