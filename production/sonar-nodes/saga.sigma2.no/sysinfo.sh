#!/usr/bin/bash
#
# Run `sonar sysinfo` and pipe its output to a network sink.

set -euf -o pipefail

# sonar_dir contains the binaries and the config file.  Other paths are defined by the config file.
sonar_dir=/cluster/shared/sonar
source $sonar_dir/sonar-config.sh

output=$($sonar_dir/sonar sysinfo)
if [[ "$output" == "" ]]; then
    exit
fi

exec curl --data-binary @- \
	   -H 'Content-Type: application/json' \
	   --netrc-file $curl_auth_file \
	   --retry 11 --retry-connrefused \
           $upload_address/sysinfo?cluster=$cluster <<< $output
