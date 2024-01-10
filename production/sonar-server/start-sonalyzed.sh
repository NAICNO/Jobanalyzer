#!/bin/bash

# This script is currently run by cron at boot time (see sonar-server.cron) to start Jobanalyzer's
# `sonalyzed` server, which runs sonalyze locally on behalf of a remote client in response to a GET.

sonar_dir=${sonar_dir:-$HOME/sonar}
source $sonar_dir/server-config

data_dir=$sonar_dir/data
mkdir -p $data_dir

pidfile=$sonar_dir/sonalyzed.pid
rm -f $pidfile
$sonar_dir/sonalyzed -jobanalyzer-path $sonar_dir -port $sonalyzed_port -password-file $sonalyzed_auth_file &
sonalyzed_pid=$!
echo $sonalyzed_pid > $pidfile
echo "SONALYZED RUNNING, PID=$sonalyzed_pid"
