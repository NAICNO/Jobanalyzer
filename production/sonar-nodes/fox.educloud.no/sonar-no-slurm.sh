#!/bin/bash
#
# Run sonar and capture its output in a file appropriate for the current time and system.
# This is for systems without slurm, ie interactive and login nodes.

set -euf -o pipefail

sonar_dir=/cluster/var/sonar
sonar_bin_dir=$sonar_dir/bin
sonar_data_dir=$sonar_dir/data
output_dir=${sonar_data_dir}/$(date +'%Y/%m/%d')

mkdir -p ${output_dir}

# TODO: It's not obvious that --rollup is right for Jobanalyzer

$sonar_bin_dir/sonar ps \
		     --exclude-system-jobs \
		     --exclude-commands=bash,ssh,zsh,tmux,systemd \
		     --min-cpu-time=60 \
		     --rollup \
		     --batchless \
		     --lockdir=/var/tmp \
		     >> ${output_dir}/${HOSTNAME}.csv
