# There is an -hourly file here but that is not a valid keyword as of now.
output=$($LOGINFO hostnames .)
CHECK loginfo_basic '["host1","host3"]' "$output"

output=$($LOGINFO hostnames yossarian 2>&1)
exitcode=$?
CHECK_ERR loginfo_no_dir $exitcode "$output" 'no such file or directory'

output=$($LOGINFO hostnames ..)
CHECK loginfo_empty '[]' "$output"
