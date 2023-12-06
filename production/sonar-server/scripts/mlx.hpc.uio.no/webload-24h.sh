#!/usr/bin/bash

# Analysis job to run on one node every 24h.  This job generates the monthly and quarterly load
# reports for the nodes.

set -euf -o pipefail

cluster=mlx.hpc.uio.no
sonar_dir=$HOME/sonar
data_dir=$sonar_dir/data/$cluster
report_dir=$sonar_dir/reports/$cluster
script_dir=$sonar_dir/scripts/$cluster

mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/ml-nodes.json \
		      -output-path $report_dir \
		      -data-path $data_dir \
		      -with-downtime \
		      -tag monthly \
		      -daily \
		      -from 30d

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/ml-nodes.json \
		      -output-path $report_dir \
		      -data-path $data_dir \
		      -with-downtime \
		      -tag quarterly \
		      -daily \
		      -from 90d
