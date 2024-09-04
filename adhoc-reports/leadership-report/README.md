These scripts were used to generate some tables for the prototype NRIS Leadership Meeting report.

IMPORTANT SETUP INSTRUCTION:

  Before running be sure to cd to ../../code and run install.sh to build and install the binaries in
  ~/go/bin.  Otherwise you risk using stale code.

PROGRAMS:

  lm-slurm-heatmaps.bash generates all the txt files, which are heat maps for various kinds of jobs.
  There's no sense in rerunning this as it stands, as the date ranges in the script are fixed and
  the outputs will be the same, but the dates can be changed.

  bad-gpu-heatmap.py prints a 5x5 heatmap of job counts that use varying amounts of GPU and GPU
  memory on the Fox GPUs.

  bad-gpu-jobs.py prints a list of commands(!) that are run on Fox GPUs without using much GPU.
  This is mostly a PoC, and the report can be adapted easily to print other things, such as jobs,
  users, and accounts.
