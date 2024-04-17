# Misc tests for config files

# The Go and Rust implementations have diverging error messages, hence a little complexity here.

# The data for ML4 says we have 20 cores, and the first input record says 1714.2% CPU used, so that
# accounts for the 86% peak (really 85.71%).  However the cputime difference between the two records
# is tiny, so the average is much lower.
output=$($SONALYZE jobs --user einarvid --fmt=csv,job,rcpu --config-file good-config.json -- dummy-data.csv)
CHECK good_config "1269178,43,86" "$output"

# Duplicate record for ml1, they are identical, doesn't matter - it's an error
output=$($SONALYZE jobs --user einarvid --fmt=csv,job,rcpu --config-file dup-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR dup_host_config $exitcode "$output" '(info for host .* already defined)|(Duplicate host name in config)'

# Missing field for "cpu_cores"
output=$($SONALYZE jobs --user - --fmt=csv,job,rcpu --config-file nocpu-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR nocpu_host_config $exitcode "$output" "(Field 'cpu_cores' must be present)|(Zero or missing 'cpu_cores')"

# Can't open file
output=$($SONALYZE jobs --user - --fmt=csv,job,rcpu --config-file nonexistent-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR nocpu_host_config $exitcode "$output" "[Nn]o such file or directory"

# Bad field value for "cpu_cores"
output=$($SONALYZE jobs --user - --fmt=csv,job,rcpu --config-file badcpu-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR nocpu_host_config $exitcode "$output" "(Field 'cpu_cores' must have unsigned integer value)|(json: cannot unmarshal number.*cpu_cores)"

# Syntax error
output=$($SONALYZE jobs --user - --config-file bad-syntax-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR badsyntax_host_config $exitcode "$output" "(ERROR.*at line)|(unmarshaling.*invalid character)"

# Smoketest
output=$($SONALYZE jobs --user einarvid --fmt=csv,job,rcpu  --config-file good-config-v2.json -- dummy-data.csv 2>&1)
CHECK good_config_v2 "1269178,43,86" "$output"
