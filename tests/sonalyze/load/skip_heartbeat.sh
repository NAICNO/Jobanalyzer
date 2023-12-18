# In the end it is correct for `load` to *not* skip heartbeat records.  If the heartbeat record is
# skipped there will be no output; if it is not skipped there will be some.

output=$($SONALYZE load -u _sonar_ -f 2023-10-05 -t 2023-10-05 --fmt=csv,time,cpu -- skip_heartbeat.csv | grep 12:00)
CHECK skip_heartbeat "12:00,0" "$output"
