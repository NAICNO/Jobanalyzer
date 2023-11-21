#!/bin/bash

# This script is currently run by cron at boot time (see sonar-server.cron) to start Jobanalyzer's
# `infiltrate` server, which receives data from all logging nodes.

sonar_dir=${sonar_dir:-$HOME/sonar}
source $sonar_dir/server-config

data_dir=$sonar_dir/data
mkdir -p $data_dir

pidfile=$sonar_dir/infiltrate.pid
rm -f $pidfile
$sonar_dir/infiltrate -data-path $data_dir -port $infiltrate_port -auth-file $infiltrate_auth_file &
infiltrate_pid=$!
echo $infiltrate_pid > $pidfile
echo "INFILTRATE RUNNING, PID=$infiltrate_pid"
