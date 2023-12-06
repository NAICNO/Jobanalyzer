#!/usr/bin/bash

# Analysis job to run on one node every 5m.  This job generates the `minutely` load reports for the
# nodes and uploads them to the server.
#
# It also generates a file of hostnames and uploads that.

# This combines webload-1h.sh with upload-data.sh.
#
# NOTE!  If upload logic needs to change here, also consider upload-data.sh.

# We need globbing, stay away from -f
set -eu -o pipefail

cluster=mlx.hpc.uio.no
sonar_dir=$HOME/sonar
data_dir=$sonar_dir/data/$cluster
report_dir=$sonar_dir/reports/$cluster
script_dir=$sonar_dir/scripts/$cluster
state_dir=$sonar_dir/state/$cluster

mkdir -p $state_dir
mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/ml-nodes.json \
		      -data-path $data_dir \
		      -output-path $report_dir \
		      -with-downtime \
		      -tag minutely \
		      -none

$sonar_dir/naicreport at-a-glance \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/ml-nodes.json \
		      -data-path $data_dir \
		      -state-path $state_dir \
		      -tag "ML Nodes" \
		      > $report_dir/ml-at-a-glance.json

$sonar_dir/naicreport hostnames \
		      $report_dir \
		      > $report_dir/ml-hostnames.json

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $report_dir/*-minutely.json
chmod go+r $report_dir/ml-at-a-glance.json
chmod go+r $report_dir/ml-hostnames.json

source $sonar_dir/upload-config.sh

upload_files="$report_dir/*-minutely.json $report_dir/ml-hostnames.json $report_dir/ml-at-a-glance.json"
if [[ $# -eq 0 || $1 != NOUPLOAD ]]; then
    # StrictHostKeyChecking has to be disabled here because this is not an interactive script, and
    # the VM has not been configured to respond in such a way that the value in known_hosts will
    # bypass the interactive prompt.
    scp -C -q -o StrictHostKeyChecking=no -i $IDENTITY_FILE_NAME \
	$upload_files \
	$WWWUSER_AND_HOST:$WWWUSER_UPLOAD_PATH
else
    echo $upload_files
fi

