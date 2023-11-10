#!/bin/bash
#
# Run sonar and capture its output in a file appropriate for the current time and system.

set -euf -o pipefail

sonar_bin_dir=/cluster/var/sonar/bin
sonar_data_dir=/cluster/var/sonar/data

path=$(date +'%Y/%m/%d')
output_dir=${sonar_data_dir}/${path}
mkdir -p ${output_dir}

# Fox has a job queue, so do not use --batchless
# TODO: It's not obvious that --rollup is right for Jobanalyzer

$sonar_bin_dir/sonar ps --exclude-system-jobs --exclude-commands=bash,ssh,zsh,tmux,systemd --min-cpu-time=60 --rollup --lockdir=/var/tmp >> ${output_dir}/${HOSTNAME}.csv
