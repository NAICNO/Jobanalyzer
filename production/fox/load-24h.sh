#!/usr/bin/bash

# Analysis job to run on one node every 24h.  This job generates the monthly and quarterly load
# reports for the nodes.

set -euf -o pipefail

sonar_dir=$HOME/sonar
data_dir=$sonar_dir/data
report_dir=$sonar_dir/output

mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $sonar_dir/fox.json \
		      -data-path $data_dir \
		      -with-downtime \
		      -tag monthly \
		      -daily \
		      -from 30d \
		      -output-path $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $sonar_dir/fox.json \
		      -data-path $data_dir \
		      -with-downtime \
		      -tag quarterly \
		      -daily \
		      -from 90d \
		      -output-path $report_dir
