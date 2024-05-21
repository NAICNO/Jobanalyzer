#!/usr/bin/bash
#
# Run `sonar ps` and pipe its output to a network sink.

set -euf -o pipefail

sonar_dir=/cluster/var/sonar
source $sonar_dir/sonar-config.sh

# --batchless is for systems without a job queue
#
# --lockdir holds a lock file that, if present, will cause sonar to exit.  The use of a lock file
#   was requested by the Fox admins.
#
# The structure here is deliberately what it is so that if sonar finds a lock file to be present, or
# fails for any other reason, then we will not create an uploader process.

output=$($sonar_dir/sonar ps \
                          --exclude-system-jobs \
                          --exclude-commands=bash,ssh,zsh,tmux,systemd \
                          --min-cpu-time=60 \
                          --batchless \
                          --lockdir=$lockdir)
if [[ "$output" == "" ]]; then
    exit
fi

sleep $(( RANDOM % $upload_window ))
exec curl --data-binary @- \
           -H 'Content-Type: text/csv' \
           --netrc-file $curl_auth_file \
           --retry 11 --retry-connrefused \
           $upload_address/sonar-freecsv?cluster=$cluster <<< $output
