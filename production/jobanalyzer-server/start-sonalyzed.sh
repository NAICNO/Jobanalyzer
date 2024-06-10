#!/bin/bash

# This script is currently run by cron at boot time (see jobanalyzer.cron) to start Jobanalyzer's
# `sonalyze daemon` (previously known as `sonalyzed`) server, which runs sonalyze locally on behalf
# of a remote client in response to a GET or POST.

sonar_dir=${sonar_dir:-$HOME/sonar}
source $sonar_dir/server-config

data_dir=$sonar_dir/data
mkdir -p $data_dir

pidfile=$sonar_dir/sonalyzed.pid
rm -f $pidfile
$sonar_dir/sonalyze daemon \
    -jobanalyzer-dir $sonar_dir \
    -port $sonalyzed_port \
    -match-user-and-cluster \
    -analysis-auth $sonalyzed_analysis_auth_file \
    -upload-auth $sonalyzed_upload_auth_file &
sonalyzed_pid=$!
echo $sonalyzed_pid > $pidfile
echo "SONALYZED RUNNING, PID=$sonalyzed_pid"
