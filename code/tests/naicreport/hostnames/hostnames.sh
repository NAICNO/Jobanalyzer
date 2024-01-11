# There is an -hourly file here but that is not a valid keyword as of now.
output=$($NAICREPORT hostnames .)
CHECK hostnames_basic '["host1","host3"]' "$output"

output=$($NAICREPORT hostnames yossarian 2>&1)
exitcode=$?
CHECK_ERR hostnames_no_dir $exitcode "$output" 'no such file or directory'

output=$($NAICREPORT hostnames ..)
CHECK hostnames_empty '[]' "$output"
