#!/usr/bin/bash

# Analysis job to run on one node every 24h.  This job generates the monthly and quarterly load
# reports for the nodes.

set -euf -o pipefail

cluster=saga.sigma2.no

sonar_dir=${sonar_dir:-$HOME/sonar}
data_dir=$sonar_dir/data/$cluster
report_dir=$sonar_dir/reports/$cluster
script_dir=$sonar_dir/scripts/$cluster

mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -report-dir $report_dir \
		      -data-dir $data_dir \
		      -with-downtime 5 \
		      -tag monthly \
		      -daily \
		      -from 30d

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -report-dir $report_dir \
		      -data-dir $data_dir \
		      -with-downtime 5 \
		      -tag quarterly \
		      -daily \
		      -from 90d
