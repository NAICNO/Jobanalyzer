#!/usr/bin/bash

# Analysis job to run on the analysis host every 1h.  This job generates the hourly and daily load
# reports for the nodes.

set -euf -o pipefail

cluster=fox.educloud.no

sonar_dir=${sonar_dir:-$HOME/sonar}
data_dir=$sonar_dir/data/$cluster
report_dir=$sonar_dir/reports/$cluster
script_dir=$sonar_dir/scripts/$cluster

mkdir -p $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -data-dir $data_dir \
		      -with-downtime \
		      -tag daily \
		      -hourly \
		      -report-dir $report_dir

$sonar_dir/naicreport load \
		      -sonalyze $sonar_dir/sonalyze \
		      -config-file $script_dir/$cluster-config.json \
		      -data-dir $data_dir \
		      -with-downtime \
		      -tag weekly \
		      -hourly \
		      -from 7d \
		      -report-dir $report_dir

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $script_dir/$cluster-config.json \
                      -data-dir $data_dir \
                      -tag fox-cpu-weekly \
                      -hourly \
                      -from 7d \
                      -group 'c1-[5-29]' \
                      -report-dir $report_dir

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $script_dir/$cluster-config.json \
                      -data-dir $data_dir \
                      -tag fox-gpu-weekly \
                      -hourly \
                      -from 7d \
                      -group 'gpu-[1-10]' \
                      -report-dir $report_dir

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $script_dir/$cluster-config.json \
                      -data-dir $data_dir \
                      -tag fox-int-weekly \
                      -hourly \
                      -from 7d \
                      -group 'int-[1-4]' \
                      -report-dir $report_dir

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $script_dir/$cluster-config.json \
                      -data-dir $data_dir \
                      -tag fox-login-weekly \
                      -hourly \
                      -from 7d \
                      -group 'login-[1-4]' \
                      -report-dir $report_dir
