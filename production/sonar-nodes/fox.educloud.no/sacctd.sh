#!/usr/bin/bash
#
# Run `sacctd` and pipe its output to a network sink.  This should only be run on one node!
# And about once an hour.

set -euf -o pipefail

sonar_dir=/cluster/var/sonar
source $sonar_dir/sonar-config.sh

output=$($sonar_dir/sacctd)
if [[ "$output" == "" ]]; then
    exit
fi

exec curl --data-binary @- \
	   -H 'Content-Type: text/csv' \
	   --netrc-file $curl_auth_file \
	   --retry 11 --retry-connrefused \
           "$upload_address/add?slurm-sacct=xxxxxtruexxxxx&cluster=$cluster" <<< $output
