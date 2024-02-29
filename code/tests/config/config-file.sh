# Misc tests for config files

# The data for ML4 says we have 20 cores, and the first input record says 1714.2% CPU used, so that
# accounts for the 86% peak (really 85.71%).  However the cputime difference between the two records
# is tiny, so the average is much lower.
output=$($SONALYZE jobs -ueinarvid --fmt=csv,job,rcpu --config-file good-config.json -- dummy-data.csv)
CHECK good_config "1269178,43,86" "$output"

# Duplicate record for ml1, they are identical, doesn't matter - it's an error
output=$($SONALYZE jobs -ueinarvid --fmt=csv,job,rcpu --config-file dup-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR dup_host_config $exitcode "$output" 'info for host .* already defined'

# Missing field for "cpu_cores"
output=$($SONALYZE jobs -u- --fmt=csv,job,rcpu --config-file nocpu-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR nocpu_host_config $exitcode "$output" "Field 'cpu_cores' must be present"

# Can't open file
output=$($SONALYZE jobs -u- --fmt=csv,job,rcpu --config-file nonexistent-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR nocpu_host_config $exitcode "$output" "No such file or directory"

# Bad field value for "cpu_cores"
output=$($SONALYZE jobs -u- --fmt=csv,job,rcpu --config-file badcpu-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR nocpu_host_config $exitcode "$output" "Field 'cpu_cores' must have unsigned integer value"
