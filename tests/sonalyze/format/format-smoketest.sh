# This just tests that --fmt=help works and produces at least the syntax help.
output=$($SONALYZE parse --fmt=help | grep -e "--fmt")
CHECK format_help "  --fmt=(field|alias|control),..." "$output"

output=$($SONALYZE parse --from 2023-10-04 --fmt=csv,host,user,job,gpus -- format-smoketest.csv)
CHECK format_csv "ml4.hpc.uio.no,einarvid,1269178,none" "$output"

output=$($SONALYZE parse --from 2023-10-04 --fmt=csvnamed,host,user,job,gpus -- format-smoketest.csv)
CHECK format_csvnamed "host=ml4.hpc.uio.no,user=einarvid,job=1269178,gpus=none" "$output"

# Everything is a string to `parse`.
output=$($SONALYZE parse --from 2023-10-04 --fmt=json,host,user,job,gpus -- format-smoketest.csv)
CHECK format_json '[{"host":"ml4.hpc.uio.no","user":"einarvid","job":"1269178","gpus":"none"}]' "$output"
