#!/usr/bin/bash
#
# Run sonar and pipe its output to a network sink.
#
# The upload window is set to 280 seconds so that the upload is almost certain to be done before
# sonar runs the next time, assuming a 5-minute interval for sonar runs.  Correctness does not
# depend on that - data can arrive at the server in any order, so long as they are tagged with the
# correct timestamp - but it's nice to not have more processes running concurrently than necessary.

set -euf -o pipefail

sonar_dir=$HOME/sonar
target_addr=http://158.39.48.160:8086
window=280

# --batchless is for systems without a job queue

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
	  -auth-file ~/.ssh/exfil-auth.txt

