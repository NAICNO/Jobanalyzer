#!/usr/bin/bash
#
# Run `sonar sysinfo` and pipe its output to a network sink.

set -euf -o pipefail

# sonar_dir contains the binaries and the config file.  Other paths are defined by the config file.
sonar_dir=/cluster/shared/sonar
source $sonar_dir/sonar-config.sh

outputfile=$(mktemp --tmpdir sonarXXXXXX)
trap "rm -f $outputfile" EXIT
trap "rm -f $outputfile; exit" ERR HUP TERM INT

$sonar_dir/sonar sysinfo > $outputfile

if [[ $(wc -l < $outputfile) == 0 ]]; then
    exit
fi

# This should have been an exec but then the trap won't work and the temp file will be left
# sitting around.
$curl_binary --data-binary @- \
	   -H 'Content-Type: application/json' \
           --netrc-file $curl_auth_file \
           --retry 11 --retry-connrefused \
           $upload_address/sysinfo?cluster=$cluster < $outputfile
