output=$($SONALYZE parse --fmt=job,gpu_status -- parse_gpufail.csv | sort)
CHECK gpufail "1269178,1
1269179,0" "$output"
