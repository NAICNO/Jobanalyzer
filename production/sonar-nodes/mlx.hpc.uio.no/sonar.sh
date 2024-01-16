#!/usr/bin/bash
#
# Run sonar and pipe its output to a network sink.

set -euf -o pipefail

# The sonar and exfiltrate binaries must be in this directory.
sonar_dir=$HOME

# The server receiving data
target_addr=http://naic-monitor.uio.no:1553
#target_addr=http://158.39.48.160:8086

# Must have a single username:password line, known to the receiving server
auth_file=$sonar_dir/exfil-auth.txt

# The upload window is set to 280 seconds so that the upload is almost certain to be done before
# sonar runs the next time, assuming a 5-minute interval for sonar runs.  Correctness does not
# depend on that - data can arrive at the server in any order, so long as they are tagged with the
# correct timestamp - but it's nice to not have more processes running concurrently than necessary.
window=280

# --batchless is for systems without a job queue
#
# --rollup merges processes with the same command line within the same job, it may or may not be
#   right for subsequent analysis.  It's possibly more pertinent to HPC (MPI) jobs than typical ML
#   jobs.

$sonar_dir/sonar ps \
		 --exclude-system-jobs \
		 --exclude-commands=bash,ssh,zsh,tmux,systemd \
		 --min-cpu-time=60 \
		 --batchless \
		 --rollup \
    | $sonar_dir/exfiltrate \
	  -cluster mlx.hpc.uio.no \
	  -window $window \
	  -source sonar/csvnamed \
	  -output json \
	  -target $target_addr \
	  -auth-file $auth_file

