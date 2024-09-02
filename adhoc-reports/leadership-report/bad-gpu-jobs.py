#!/usr/bin/python3
#
# See README.md for setup instructions.
#
# Example report: noninteractive programs that ran on the GPU nodes for at least 2 hours and used
# less than 25% of its allocated GPU and less than 25% of its allocated GPU memory (at peak).  This
# is pretty arbitrary!
#
# This is a good example of how sonar and slurm data are joined.  One would think that the job
# numbers to include or exclude computed in the first step could be passed to sonalyze in the second
# step, but curl has limits on command line / url parameter length.

import os, subprocess, sys, re
from sonalyze import sonalyze

# Exclude gpu-3 because it is dedicated to interactive work and has no Slurm jobs.  Maybe gpu-9 also?
cluster = "fox"
hostglob = "gpu-[2,3,4-13]"
from_date = "2024-03-01"
to_date = "2024-05-31"

####################################################################################################
# Find job IDs of all noninteractive slurm jobs.  Here we don't filter TIMEOUT and CANCELLED jobs,
# nor do we use -some-gpu, and we use -all to be maximally inclusive.

sacct_output = sonalyze("sacct", cluster,
                        ["-fmt", "awk,JobID,User,Account,State,JobName",
                         "-from", from_date,
                         "-to", to_date,
                         "-all",
                         "-host", hostglob])
noninteractive_jobs={}
interactive=re.compile(r"interactive|OOD|ood")
for l in sacct_output:
    if interactive.search(l) != None:
        continue
    fields=l.split(" ")
    noninteractive_jobs[fields[0]] = True

#print("{0} sacct records, {1} noninteractive left".format(len(sacct_output), len(noninteractive_jobs)))

####################################################################################################
# Slurp and filter job data.  Only look at jobs that ran at least 2 hours (elapsed) since those are
# considered somewhat "serious".

jobs_output = sonalyze("jobs", cluster,
                       ["-from", from_date,
                        "-to", to_date,
                        "-u", "-",
                        "-min-runtime", "2h",
                        "-host", hostglob,
                        "-fmt", "awk,job,sgpu,sgpumem,user,cmd"])
all_cmds={}
jobs={}
for l in jobs_output:
    # We want noninteractive jobs whose gpu and gpumem peaks both fall below 25%
    # fields are jobid, sgpu-avg, sgpu-peak, sgpumem-avg, sgpumem-peak, user, cmd
    fields=l.split(" ")
    jobs[fields[0]]=fields
    if fields[0] in noninteractive_jobs and int(fields[2]) < 25 and int(fields[4]) < 25:
        for c in fields[6].split(","):
            if c in all_cmds:
                all_cmds[c].append(fields[0])
            else:
                all_cmds[c] = [fields[0]]

for c in all_cmds:
    print(c + " " + str(len(all_cmds[c])))
