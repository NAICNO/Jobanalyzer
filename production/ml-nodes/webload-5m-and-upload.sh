#!/usr/bin/env bash

# Analysis job to run on one node every 5m.  This job generates the
# `minutely` load reports for the nodes and uploads them to the server.

# This combines webload-1h.sh with upload-data.sh.
#
# NOTE!  If upload logic needs to change here, also consider upload-data.sh.

# We need globbing, stay away from -f
set -eu -o pipefail

sonar_dir=$HOME/sonar
data_path=$sonar_dir/data
output_path=$sonar_dir/data/load-reports
load_report_path=$output_path

mkdir -p $output_path

common_options="--sonalyze $sonar_dir/sonalyze --config-file $sonar_dir/ml-nodes.json --output-path $output_path --data-path $data_path"
$sonar_dir/naicreport ml-webload $common_options --tag minutely --none

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $load_report_path/*-minutely.json

# StrictHostKeyChecking has to be disabled here because this is not an interactive script,
# and the VM has not been configured to respond in such a way that the value in known_hosts
# will bypass the interactive prompt.
scp -q -o StrictHostKeyChecking=no -i $sonar_dir/ubuntu-vm.pem $load_report_path/*-minutely.json ubuntu@158.39.48.160:/var/www/html/output
