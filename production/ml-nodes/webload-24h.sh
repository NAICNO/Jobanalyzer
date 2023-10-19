#!/usr/bin/env bash

# Analysis job to run on one node every 24h.  This job generates the
# monthly and quarterly load reports for the nodes.

set -euf -o pipefail

sonar_dir=$HOME/sonar
data_path=$sonar_dir/data
output_path=$sonar_dir/data/load-reports

mkdir -p $output_path

common_options="--sonalyze $sonar_dir/sonalyze --config-file $sonar_dir/ml-nodes.json --output-path $output_path --data-path $data_path --with-downtime"
$sonar_dir/naicreport ml-webload $common_options --tag monthly --daily --from 30d
$sonar_dir/naicreport ml-webload $common_options --tag quarterly --daily --from 90d
