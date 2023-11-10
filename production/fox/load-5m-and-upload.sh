#!/usr/bin/bash

# Analysis job to run on one node every 5m.  This job generates the `minutely` load reports for the
# nodes and uploads them to the server.
#
# It also generates a file of hostnames and uploads that.

# This combines load-1h.sh with upload-data.sh.
#
# NOTE!  If upload logic needs to change here, also consider upload-data.sh.

# We need globbing, stay away from -f
set -eu -o pipefail

sonar_dir=$HOME/sonar
sonar_bin=$sonar_dir
data_dir=$sonar_dir/data
output_dir=$sonar_dir/output

mkdir -p $output_dir

naicreport_options="--sonalyze $sonar_dir/sonalyze --config-file $sonar_dir/fox.json --data-path $data_dir"

$sonar_bin/naicreport load $naicreport_options --output-path $output_dir --with-downtime --tag minutely --none
$sonar_bin/naicreport at-a-glance $naicreport_options --state-path $data_dir -tag "Fox" > $output_dir/fox-at-a-glance.json

$sonar_bin/loginfo hostnames $output_dir > $output_dir/fox-hostnames.json

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $output_dir/*-minutely.json
chmod go+r $output_dir/fox-at-a-glance.json
chmod go+r $output_dir/fox-hostnames.json

source $sonar_dir/upload-config.sh

# StrictHostKeyChecking has to be disabled here because this is not an interactive script,
# and the VM has not been configured to respond in such a way that the value in known_hosts
# will bypass the interactive prompt.
if [[ $# -eq 0 || $1 != NOUPLOAD ]]; then
    scp -C -q -o StrictHostKeyChecking=no -i $IDENTITY_FILE_NAME \
	$output_dir/*-minutely.json $output_dir/fox-hostnames.json $output_dir/fox-at-a-glance.json \
	$WWWUSER_AND_HOST:$WWWUSER_UPLOAD_PATH
fi

