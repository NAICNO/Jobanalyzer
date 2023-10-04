# If the heartbeat record is not properly skipped there will be multiple lines of output; if it is
# skipped there will be none (because no output is produced for zero records).

output=$($SONALYZE load -u _sonar_ -f 2023-10-05 -t 2023-10-05 --fmt=csv,time -- skip_heartbeat.csv)
CHECK skip_heartbeat "" "$output"
