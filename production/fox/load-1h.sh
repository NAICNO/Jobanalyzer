#!/usr/bin/bash

# Analysis job to run on one node every 1h.  This job generates the hourly and daily load reports
# for the nodes.

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
		      -tag daily \
		      -hourly \
		      -output-path $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $sonar_dir/fox.json \
		      -data-path $data_dir \
		      -with-downtime \
		      -tag weekly \
		      -hourly \
		      -from 7d \
		      -output-path $report_dir
