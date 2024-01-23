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
sonar_secrets_dir=$sonar_bin_dir/secrets

# The server receiving the data.  For HTTPS you need -ca-cert below, for HTTP remove that arg.
target_addr=https://naic-monitor.uio.no:1697
#target_addr=http://naic-monitor.uio.no:1553

# Must have a single username:password line, known to the receiving server
auth_file=$sonar_secrets_dir/exfil-auth.txt

# For HTTPS only - goes with the -ca-cert line below
cert_file=$sonar_secrets_dir/exfil-ca.crt

# The upload window is set to 280 seconds so that the upload is almost certain to be done before
# sonar runs the next time, assuming a 5-minute interval for sonar runs.  Correctness does not
# depend on that - data can arrive at the server in any order, so long as they are tagged with the
# correct timestamp - but it's nice to not have more processes running concurrently than necessary.
window=240

# The canonical name of the cluster
cluster=fox.educloud.no

# Fox has a job queue, so do not use --batchless
# TODO: It's not obvious that --rollup is right for Jobanalyzer

$sonar_bin_dir/sonar ps \
		 --exclude-system-jobs \
		 --exclude-commands=bash,ssh,zsh,tmux,systemd \
		 --min-cpu-time=60 \
		 --rollup \
		 --lockdir=/var/tmp \
    | $sonar_bin_dir/exfiltrate \
	  -cluster $cluster \
	  -window $window \
	  -source sonar/csvnamed \
	  -output json \
	  -auth-file $auth_file \
	  -target $target_addr \
	  -ca-cert $cert_file
