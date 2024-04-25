output=$($SONALYZE parse --fmt=all -- parse.csv)
CHECK parse_all "0.7.0,2023-10-04 07:40,ml4.hpc.uio.no,64,0,einarvid,0,1269178,python3,1714.2,261,0,none,0,0,0,0,10192,69,0" "$output"

output=$($SONALYZE parse -- parse.csv)
CHECK parse_default "1269178,einarvid,python3" "$output"

output=$($SONALYZE parse --fmt=json,all -- empty_input.csv)
CHECK parse_json_empty "[]" "$output"

output=$($SONALYZE parse --fmt=json,all,nodefaults -- parse.csv)
CHECK parse_json '[{"version":"0.7.0","localtime":"2023-10-04 07:40","host":"ml4.hpc.uio.no","cores":"64","memtotal":"0","user":"einarvid","job":"1269178","cmd":"python3","cpu_pct":"1714.2","mem_gb":"261","res_gb":"0","cputime_sec":"10192","rolledup":"69"}]' "$output"
