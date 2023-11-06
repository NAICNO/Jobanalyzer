# JSON "empty" output is always an empty array
output=$($SONALYZE jobs --fmt=json,job,user -- empty-file.csv)
CHECK jobs_empty_json "[]" "$output"

# CSV "empty" output is truly empty, b/c noheader by default
output=$($SONALYZE jobs --fmt=csv,job,user -- empty-file.csv)
CHECK jobs_empty_csv "" "$output"

# There should still be a header printed for empty output, though I guess this is debatable.
output=$($SONALYZE jobs --fmt=job,user -- empty-file.csv)
CHECK jobs_empty_fixed "job  user" "$output"
