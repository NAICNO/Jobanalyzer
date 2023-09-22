#!/usr/bin/env bash
#
# Run sonar and capture its output in a file appropriate for the current time and system.

set -euf -o pipefail

sonar_dir=$HOME/sonar
sonar_data_dir=$sonar_dir/data

year=$(date +'%Y')
month=$(date +'%m')
day=$(date +'%d')

output_directory=${sonar_data_dir}/${year}/${month}/${day}
mkdir -p ${output_directory}

# --batchless is for systems without a job queue

$sonar_dir/sonar ps --exclude-system-jobs --exclude-commands=bash,ssh,zsh,tmux,systemd --min-cpu-time=60 --batchless --rollup >> ${output_directory}/${HOSTNAME}.csv
