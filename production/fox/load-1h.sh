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

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $sonar_dir/fox.json \
                      -data-path $data_dir \
                      -tag fox-cpu-weekly \
                      -hourly \
                      -from 7d \
                      -cluster 'c1-[5-29]' \
                      -output-path $report_dir

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $sonar_dir/fox.json \
                      -data-path $data_dir \
                      -tag fox-gpu-weekly \
                      -hourly \
                      -from 7d \
                      -cluster 'gpu-[1-9]' \
                      -output-path $report_dir

$sonar_dir/naicreport load \
                      -sonalyze $sonar_dir/sonalyze \
                      -config-file $sonar_dir/fox.json \
                      -data-path $data_dir \
                      -tag fox-int-weekly \
                      -hourly \
                      -from 7d \
                      -cluster 'int-[1-4]' \
                      -output-path $report_dir
