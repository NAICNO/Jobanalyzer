#!/usr/bin/bash
#
# Run `sonar ps` and pipe its output to a network sink.

set -euf -o pipefail

sonar_dir=/cluster/shared/sonar
source $sonar_dir/sonar-config.sh

# --rollup merges processes with the same command line within the same job, it may or may not be
#   right for subsequent analysis.  It's possibly more pertinent to HPC (MPI) jobs than typical ML
#   jobs.
#
# TODO: It's not obvious that --rollup is right for Jobanalyzer since sometimes we want to view
# separate processes separately.
#
# The structure here is deliberately what it is so that if sonar finds a lock file to be present, or
# fails for any other reason, then we will not create an uploader process.

output=$($sonar_dir/sonar ps \
                          --exclude-system-jobs \
                          --exclude-commands=bash,ssh,zsh,tmux,systemd \
                          --min-cpu-time=60 \
                          --rollup )
if [[ "$output" == "" ]]; then
    exit
fi

sleep $(( RANDOM % $upload_window ))
exec curl --data-binary @- \
           -H 'Content-Type: text/csv' \
           --netrc-file $curl_auth_file \
           --retry 11 --retry-connrefused \
           $upload_address/sonar-freecsv?cluster=$cluster <<< $output
