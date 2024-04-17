# This just tests that --fmt=help works and produces at least the syntax help.
# Rust and Go have slightly different output due to the - / -- discrepancy.
# Also, Rust prints on stdout and Go on stderr (the latter to do what -h would do)
output=$($SONALYZE parse --fmt=help 2>&1 | grep -e "-fmt" | sed 's/^ *-*//g')
CHECK format_help "fmt=(field|alias|control),..." "$output"

output=$($SONALYZE parse --fmt=csv,host,user,job,gpus -- format-smoketest.csv)
CHECK format_csv "ml4.hpc.uio.no,einarvid,1269178,none" "$output"

output=$($SONALYZE parse --fmt=csvnamed,host,user,job,gpus -- format-smoketest.csv)
CHECK format_csvnamed "host=ml4.hpc.uio.no,user=einarvid,job=1269178,gpus=none" "$output"

# Everything is a string to `parse`.
output=$($SONALYZE parse --fmt=json,host,user,job,gpus -- format-smoketest.csv)
CHECK format_json '[{"host":"ml4.hpc.uio.no","user":"einarvid","job":"1269178","gpus":"none"}]' "$output"
