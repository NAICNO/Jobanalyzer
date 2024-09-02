#!/usr/bin/python3
#
# See README.md for setup instructions.
#
# Example report: heat map of peak gpu and gpu memory usage of programs that ran on the GPU nodes
# for at least 2 hours.

import os, subprocess, sys, re
from sonalyze import sonalyze
from heatmap import heatmap

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
                        "-fmt", "awk,job,sgpu,sgpumem"])

out=[]
for l in jobs_output:
    fields=l.split(" ")
    if fields[0] in noninteractive_jobs:
        out.append([fields[2], fields[4]])
                   
for l in heatmap(out, 5):
    print(l)
