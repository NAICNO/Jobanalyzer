#!/usr/bin/bash

# Analysis job to run on one node every 24h.  This job generates the
# monthly and quarterly load reports for the nodes.

set -euf -o pipefail

sonar_dir=$HOME/sonar
sonar_bin=$sonar_dir
data_dir=$sonar_dir/data
output_dir=$sonar_dir/output

mkdir -p $output_dir

common_options="--sonalyze $sonar_dir/sonalyze --config-file $sonar_dir/fox.json --output-path $output_dir --data-path $data_dir --with-downtime"
$sonar_dir/naicreport load $common_options --tag monthly --daily --from 30d
$sonar_dir/naicreport load $common_options --tag quarterly --daily --from 90d
