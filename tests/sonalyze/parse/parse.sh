output=$($SONALYZE parse --from 2023-10-04 --fmt=all -- parse.csv)
CHECK parse_all "0.7.0,2023-10-04 07:40,ml4.hpc.uio.no,64,einarvid,0,1269178,python3,1714.2,261,none,0,0,0,0,10192,69,0" "$output"

output=$($SONALYZE parse --from 2023-10-04 -- parse.csv)
CHECK parse_default "1269178,einarvid,python3" "$output"
