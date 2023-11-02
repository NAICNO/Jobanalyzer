#!/usr/bin/env bash

# Analysis job to run on one node every 5m.  This job generates the
# `minutely` load reports for the nodes and uploads them to the server.
#
# It also generates a file of hostnames and uploads that.

# This combines webload-1h.sh with upload-data.sh.
#
# NOTE!  If upload logic needs to change here, also consider upload-data.sh.

# We need globbing, stay away from -f
set -eu -o pipefail

sonar_dir=$HOME/sonar
data_path=$sonar_dir/data
output_path=$sonar_dir/data/load-reports

mkdir -p $output_path

naicreport_options="--sonalyze $sonar_dir/sonalyze --config-file $sonar_dir/ml-nodes.json --data-path $data_path"

$sonar_dir/naicreport ml-webload $naicreport_options --output-path $output_path --with-downtime --tag minutely --none
$sonar_dir/naicreport at-a-glance $naicreport_options --state-path $data_path > $output_path/at-a-glance.json

$sonar_dir/loginfo hostnames $output_path > $output_path/hostnames.json

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $output_path/*-minutely.json
chmod go+r $output_path/at-a-glance.json
chmod go+r $output_path/hostnames.json

source $sonar_dir/upload-config.sh

# StrictHostKeyChecking has to be disabled here because this is not an interactive script,
# and the VM has not been configured to respond in such a way that the value in known_hosts
# will bypass the interactive prompt.
if [[ $# -eq 0 || $1 != NOUPLOAD ]]; then
    scp -C -q -o StrictHostKeyChecking=no -i $IDENTITY_FILE_NAME \
	$output_path/*-minutely.json $output_path/hostnames.json $output_path/at-a-glance.json \
	$WWWUSER_AND_HOST:$WWWUSER_UPLOAD_PATH
fi

