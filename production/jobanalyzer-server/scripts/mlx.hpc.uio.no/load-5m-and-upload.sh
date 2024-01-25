#!/usr/bin/bash

# Analysis job to run on one node every 5m.  This job generates the `minutely` load reports for the
# nodes and uploads them to the server.
#
# It also generates a file of hostnames and uploads that.

# We need globbing, stay away from -f
set -eu -o pipefail

cluster=mlx.hpc.uio.no
abbrev=ml
tag="ML Nodes"

sonar_dir=${sonar_dir:-$HOME/sonar}
data_dir=$sonar_dir/data/$cluster
report_dir=$sonar_dir/reports/$cluster
script_dir=$sonar_dir/scripts/$cluster
state_dir=$sonar_dir/state/$cluster

mkdir -p $state_dir
mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -data-dir $data_dir \
		      -report-dir $report_dir \
		      -with-downtime \
		      -tag minutely \
		      -none

$sonar_dir/naicreport at-a-glance \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -data-dir $data_dir \
		      -state-dir $state_dir \
		      -tag "$tag" \
		      > $report_dir/$abbrev-at-a-glance.json

$sonar_dir/naicreport hostnames \
		      -report-dir $report_dir \
		      > $report_dir/$abbrev-hostnames.json

upload_files="$report_dir/*-minutely.json $report_dir/$abbrev-hostnames.json $report_dir/$abbrev-at-a-glance.json"
source $script_dir/upload-subr.sh
