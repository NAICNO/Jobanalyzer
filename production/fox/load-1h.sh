#!/usr/bin/bash

# Analysis job to run on one node every 1h.  This job generates the hourly and daily load reports
# for the nodes.

set -euf -o pipefail

sonar_dir=$HOME/sonar
sonar_bin=$sonar_dir
data_dir=$sonar_dir/data
output_dir=$sonar_dir/output

mkdir -p $output_dir

common_options="--sonalyze $sonar_dir/sonalyze --config-file $sonar_dir/fox.json --output-path $output_dir --data-path $data_dir --with-downtime"
$sonar_bin/naicreport load $common_options --tag daily --hourly
$sonar_bin/naicreport load $common_options --tag weekly --hourly --from 7d
