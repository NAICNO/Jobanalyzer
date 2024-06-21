#!/usr/bin/python3
#
# Look for heavy users of GPU resources.

import subprocess

# TODO: These are parameters
cluster = "fox"
timespan = "1w"

# TODO: Absolute paths need to be figured out somehow, but subprocess.run() does not
# deal with relative paths, it must have a cwd
auth_file = "/home/larstha/.ssh/sonalyzed-auth.txt"
sonalyze = "/home/larstha/p/Jobanalyzer/code/sonalyze/sonalyze"
remote = "https://naic-monitor.uio.no"
secperday = 86400

def heavy_users():
    # run sonalyze, capture the output
    cp = subprocess.run([sonalyze, "jobs",
                         "-remote", remote,
                         "-auth-file", auth_file,
                         "-cluster", cluster,
                         "-host", "gpu-*",
                         "-from", timespan,
                         "-u", "-",
                         "-some-gpu",
                         "-fmt", "awk,duration/sec,gputime/sec,user,job,host,cmd"],
                        capture_output=True, text=True)
    if cp.returncode != 0:
        raise RuntimeError("Sonalyze jobs failed")
    for l in cp.stdout.splitlines():
        fields=l.split(" ")

        # skip anything that ran for less than a day
        if int(fields[0]) < secperday:
            continue

        # skip anything that didn't use a full gpu-day
        if int(fields[1]) < secperday:
            continue

        # we will print this line but must annotate it if it used a full gpu for at least a day.  to
        # see this, we must generate a profile.  The gpu usage of a time step is the sum across the
        # gpu fields for the step.  we can collect these in a timeline and look for a 24h window
        # where every slot has >= 100.
        pp = subprocess.run([sonalyze, "profile",
                             "-remote", remote,
                             "-auth-file", auth_file,
                             "-cluster", cluster,
                             "-job", fields[3],
                             "-from", timespan,
                             "-fmt", "csv,gpu",
                             "-bucket", "6"],
                            capture_output=True, text=True)
        if pp.returncode != 0:
            raise RuntimeError("Sonalyze profile failed" + pp.stderr)

        # For the process, generate a timeline.
        timeline=[]
        for pl in pp.stdout.splitlines():
            pfields=pl.split(",")
            if pfields[0] == "time":
                continue
            t=pfields[0]
            sum=0
            for x in pfields[1:]:
                if x != "":
                    sum += int(x)
            timeline.append(sum)

        # With a bucket of 6, 24 hours should be 48 items.  Iterate across the array looking
        # for a run of that length.
        i=0
        mark = False
        while not mark and i < len(timeline):
            while i < len(timeline) and timeline[i] < 100:
                i+=1
            start=i
            while i < len(timeline) and timeline[i] >= 100:
                i+=1
            if i-start >= 48:
                mark = True

        if mark:
            print("*", l)
        else:
            print(l)

# TODO: Presumably some special incantation
heavy_users()
        
