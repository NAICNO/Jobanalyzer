#!/usr/bin/bash

# Analysis job to run on the analysis host every 1h.  This job generates the hourly and daily load
# reports for the nodes.

set -euf -o pipefail

cluster=mlx.hpc.uio.no
sonar_dir=${sonar_dir:-$HOME/sonar}
data_dir=$sonar_dir/data/$cluster
report_dir=$sonar_dir/reports/$cluster
script_dir=$sonar_dir/scripts/$cluster

mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -data-dir $data_dir \
		      -with-downtime 5 \
		      -tag daily \
		      -hourly \
		      -report-dir $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -data-dir $data_dir \
		      -with-downtime 5 \
		      -tag weekly \
		      -hourly \
		      -from 7d \
		      -report-dir $report_dir

$sonar_dir/naicreport load \
		     -sonalyze $sonar_dir/sonalyze \
		     -config-file $script_dir/$cluster-config.json \
		     -data-dir $data_dir \
		     -tag ml-nvidia-weekly \
		     -hourly \
		     -from 7d \
		     -group 'ml[1-3,6-9]' \
		     -report-dir $report_dir
