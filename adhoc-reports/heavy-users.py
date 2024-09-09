#!/usr/bin/python3
#
# Look for heavy users of GPU resources.  At the moment, a heavy user is:
#
# - user with jobs occupying a full GPU at full capacity for more than 24hours
# - user with jobs occupying at least 24hours worth of GPU across the lifetime of the job
#
# (there are additional criteria to be implemented.)  In the listing produced, the former case is
# annotated with a "*" in the first column.

# Usage:
#
#  heavy-users.py cluster-name [from-date [to-date]]
#
# The cluster-name is a full name or an alias, regrettably this script contains a table of valid
# names.
#
# The from and to dates are absolute dates yyyy-mm-dd or relative dates Nd (days ago), Nw (weeks
# ago).
#
# NOTES.
#
#   To run this, your ~/.ssh/sonalyzed-auth.txt must contain your identity and password for running
#   sonalyze remotely.
#
#   This must currently be run from the directory it's in, so that it can find the sonalyze binary,
#   which you must have compiled.

import os, subprocess, sys

# TODO: Very unfortunate to have this table here.

clusters = {
    "ml":"",
    "mlx":"",
    "mlx.hpc.uio.no":"",
    "fox":"gpu-*",
    "fox.educloud.no":"gpu-*",
}

secsPerDay = 86400
secsPerHour = 3600
secsPerMinute = 60

auth_file = os.path.join(os.getenv("HOME"), ".ssh/sonalyzed-auth.txt")
sonalyze = "../code/sonalyze/sonalyze"
remote = "https://naic-monitor.uio.no"
cutoff = secsPerDay
bucketing=6
buckets_per_period=12/bucketing*24

def main():
    args = sys.argv[1:]
    if len(args) == 0:
        raise RuntimeError("Cluster name required")
    cluster=args[0]
    if not (cluster in clusters):
        raise RuntimeError("Unknown cluster name " + cluster)
    from_date=""
    to_date=""
    if len(args) > 1:
        from_date=args[1]
        if len(args) > 2:
            to_date=args[2]
    heavy_gpu_users(cluster, clusters[cluster], from_date, to_date)

def heavy_gpu_users(cluster, hostglob, from_date, to_date):
    # run sonalyze, capture the output
    jobs_cmd = [sonalyze, "jobs",
                "-remote", remote,
                "-auth-file", auth_file,
                "-cluster", cluster,
                "-u", "-",
                "-some-gpu",
                "-fmt", "awk,duration/sec,gputime/sec,user,job,host,cmd"]
    if from_date != "":
        jobs_cmd += ["-from", from_date]
    if to_date != "":
        jobs_cmd += ["-to", to_date]
    if hostglob != "":
        jobs_cmd += ["-host", hostglob]

    # Indices based on the -fmt string above
    duration_ix=0
    gputime_ix=1
    user_ix=2
    job_ix=3
    host_ix=4
    cmd_ix=5

    job_proc = subprocess.run(jobs_cmd, capture_output=True, text=True, cwd=os.getcwd())
    if job_proc.returncode != 0:
        raise RuntimeError("Sonalyze jobs failed:" + job_proc.stderr)

    rows = [[">24h","User","GpuTime","GpuTime/","Host(s)","Command"],
            ["",    "",    "",       "duration"]]
    for job_line in job_proc.stdout.splitlines():
        job_fields=job_line.split(" ")

        # Skip jobs without ID, not completely clear why these happen.
        if int(job_fields[job_ix]) == 0:
            continue

        # Skip anything that ran for less than the period.
        if int(job_fields[duration_ix]) < cutoff:
            continue

        # Skip anything that didn't use a full period's worth of GPU.
        if int(job_fields[gputime_ix]) < cutoff:
            continue

        # We will print the job_line but must annotate it if it used a full gpu for at least a day.
        # To see this, we must generate a profile.  The gpu usage of a time step is the sum across
        # the gpu fields for the step.  We can collect these in a timeline and look for a 24h window
        # where every slot has >= 100.
        #
        # Filter by host here because this massively speeds up processing for large data sets.
        prof_cmd = [sonalyze, "profile",
                    "-remote", remote,
                    "-auth-file", auth_file,
                    "-cluster", cluster,
                    "-job", job_fields[job_ix],
                    "-host", job_fields[host_ix],
                    "-fmt", "csv,gpu",
                    "-bucket", str(bucketing)]
        if from_date != "":
            prof_cmd += ["-from", from_date]
        if to_date != "":
            prof_cmd += ["-to", to_date]

        prof_proc = subprocess.run(prof_cmd, capture_output=True, text=True, cwd=os.getcwd())
        if prof_proc.returncode != 0:
            raise RuntimeError("Sonalyze profile failed: " + job_line + "\n" + prof_proc.stderr)

        # For the process, generate a timeline of GPU usage.
        timeline=[]
        for prof_line in prof_proc.stdout.splitlines():
            prof_fields=prof_line.split(",")
            if prof_fields[0] == "time":
                continue
            t=prof_fields[0]
            sum=0
            for x in prof_fields[1:]:
                if x != "":
                    sum += int(x)
            timeline.append(sum)

        # TODO: Should definitely go by timestamps here, not number of buckets, but this is OK for a
        # prototype.
        i=0
        mark = " "
        while i < len(timeline):
            while i < len(timeline) and timeline[i] < 100:
                i+=1
            start=i
            while i < len(timeline) and timeline[i] >= 100:
                i+=1
            if i-start >= buckets_per_period:
                mark = "*"
                break

        # TODO: Duration might be interesting.
        rows.append([mark,
                     job_fields[user_ix],
                     formatDuration(int(job_fields[gputime_ix])),
                     str(int(int(job_fields[gputime_ix])/int(job_fields[duration_ix])*100)) + "%",
                     job_fields[host_ix],
                     job_fields[cmd_ix]])
    tabulate(rows)

def tabulate(rows):
    m = 0
    for r in rows:
        m = max(m, len(r))
    widths = list(range(0, m))
    for r in rows:
        for i in range(0, len(r)):
            widths[i] = max(widths[i], len(r[i]))
    for r in rows:
        l = ""
        for i in range(0, len(r)):
            s = r[i]
            while len(s) < widths[i]:
                s = s + " "
            s = s + "  "
            l = l + s
        print(l)

def formatDuration(secs):
    days = int(secs / secsPerDay)
    secs = secs % secsPerDay
    hours = int(secs / secsPerHour)
    secs = secs % secsPerHour
    minutes = int(secs / secsPerMinute)
    secs = secs % secsPerMinute
    s = ""
    if days > 0:
        s = s + str(days) + "d"
    if hours > 0 or days > 0:
        s = s + str(hours) + "h"
    if minutes > 0 or hours > 0 or days > 0:
        s = s + str(minutes) + "m"
    s = s + str(secs) + "s"
    return s
    

if __name__ == "__main__":
    main()
