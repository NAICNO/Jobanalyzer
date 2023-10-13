# Misc tests for config files

# The data for ML4 says we have 20 cores, and the first input record says 1714.2% CPU used, so that
# accounts for the 86% peak (really 85.71%).  However the cputime difference between the two records
# is tiny, so the average is much lower.
output=$($SONALYZE jobs -ueinarvid --from 2023-10-04 --fmt=csv,job,rcpu --config-file good-config.json -- dummy-data.csv)
CHECK good_config "1269178,43,86" "$output"

# Duplicate record for ml1, they are identical, doesn't matter - it's an error
output=$($SONALYZE jobs -ueinarvid --from 2023-10-04 --fmt=csv,job,rcpu --config-file dup-host-config.json -- dummy-data.csv 2>&1)
exitcode=$?
CHECK_ERR dup_host_config $exitcode "$output" 'info for host .* already defined'

# There are many more error conditions to check, but one is enough to check that
# error propagation at least works.
#
# TODO: Test various failure modes, at least these (see code):
#
# - open error
# - general json parse error (various classes, do we care?)
# - missing field values
# - bad data values in known fields
