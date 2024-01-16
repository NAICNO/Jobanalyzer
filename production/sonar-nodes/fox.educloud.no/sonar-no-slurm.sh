#!/usr/bin/bash
#
# Run sonar and pipe its output to a network sink.
#
# The upload window is set to 280 seconds so that the upload is almost certain to be done before
# sonar runs the next time, assuming a 5-minute interval for sonar runs.  Correctness does not
# depend on that - data can arrive at the server in any order, so long as they are tagged with the
# correct timestamp - but it's nice to not have more processes running concurrently than necessary.

set -euf -o pipefail

sonar_bin_dir=/cluster/var/sonar/bin
target_addr=http://naic-monitor.uio.no:1553
window=240

# Fox has a job queue, so do not use --batchless
# TODO: It's not obvious that --rollup is right for Jobanalyzer

$sonar_bin_dir/sonar ps \
		 --exclude-system-jobs \
		 --exclude-commands=bash,ssh,zsh,tmux,systemd \
		 --min-cpu-time=60 \
		 --rollup \
		 --batchless \
		 --lockdir=/var/tmp \
    | $sonar_bin_dir/exfiltrate \
	  -cluster fox.educloud.no \
	  -window $window \
	  -source sonar/csvnamed \
	  -output json \
	  -target $target_addr \
	  -auth-file $sonar_bin_dir/exfil-auth.txt
