#!/usr/bin/bash
#
# Run sonar and pipe its output to a network sink.

set -euf -o pipefail

# The sonar and exfiltrate binaries must be in this directory.
sonar_dir=$HOME/sonar

# The server receiving data.  For HTTPS you need -ca-cert below, for HTTP you must remove that arg.
target_addr=https://naic-monitor.uio.no

# Must have a single username:password line, known to the receiving server
auth_file=$sonar_dir/secrets/exfil-auth.txt

# For HTTPS only.  Goes with the -cert-ca argument to exfiltrate command line below.
cert_file=$sonar_dir/secrets/naic-monitor.uio.no_fullchain.crt

# The upload window is set to 280 seconds so that the upload is almost certain to be done before
# sonar runs the next time, assuming a 5-minute interval for sonar runs.  Correctness does not
# depend on that - data can arrive at the server in any order, so long as they are tagged with the
# correct timestamp - but it's nice to not have more processes running concurrently than necessary.
window=280

# The canonical name of the cluster
cluster=saga.sigma2.no

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
	  -cluster $cluster \
	  -window $window \
	  -source sonar/csvnamed \
	  -output json \
	  -auth-file $auth_file \
	  -target $target_addr \
	  -ca-cert $cert_file
