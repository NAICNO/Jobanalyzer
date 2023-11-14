#!/usr/bin/bash
#
# Run sonar and capture its output in a file appropriate for the current time and system.

set -euf -o pipefail

sonar_dir=$HOME/sonar
data_dir=$sonar_dir/data
output_dir=${data_dir}/$(date +'%Y/%m/%d')

mkdir -p ${output_dir}

# --batchless is for systems without a job queue

$sonar_dir/sonar ps \
		 --exclude-system-jobs \
		 --exclude-commands=bash,ssh,zsh,tmux,systemd \
		 --min-cpu-time=60 \
		 --batchless \
		 --rollup \
		 >> ${output_dir}/${HOSTNAME}.csv
