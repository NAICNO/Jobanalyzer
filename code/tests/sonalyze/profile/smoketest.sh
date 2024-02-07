output=$($SONALYZE profile -j1119125 -f2023-10-21 -- smoketest.csv)
CHECK profile_smoketest \
      "time              cpu   mem  gpu  gpumem  cmd
2023-10-21 08:35  3075  5    0    0       bwa
2023-10-21 08:40  3094  5    0    0       bwa
2023-10-21 08:45  3092  5    0    0       bwa
                  11    259  0    0       samtools
2023-10-21 08:50  3078  5    0    0       bwa
                  12    259  0    0       samtools
2023-10-21 08:55  3093  5    0    0       bwa
                  11    259  0    0       samtools
2023-10-21 09:00  3080  5    0    0       bwa
                  14    262  0    0       samtools
2023-10-21 09:05  3091  5    0    0       bwa
                  11    262  0    0       samtools
2023-10-21 09:10  3091  5    0    0       bwa
                  10    262  0    0       samtools
2023-10-21 09:15  3077  5    0    0       bwa
                  13    262  0    0       samtools" \
      "$output"

output=$($SONALYZE profile -j1119125 -f2023-10-21 --fmt=csv,cpu -- smoketest.csv)
CHECK profile_smoketest_csv \
      "time,bwa (1119426),samtools (1119428)
2023-10-21 08:35,3075,
2023-10-21 08:40,3094,
2023-10-21 08:45,3092,11
2023-10-21 08:50,3078,12
2023-10-21 08:55,3093,11
2023-10-21 09:00,3080,14
2023-10-21 09:05,3091,11
2023-10-21 09:10,3091,10
2023-10-21 09:15,3077,13" \
      "$output"

# Easier to generate JSON than to write it by hand...

# T timestamp point ...
T() {
    local v
    local first
    first=1
    v="{\"time\":\"$1\",\"job\":1119125,\"points\":["
    shift
    while [[ $1 != "" ]] ; do
	if [[ $first -ne 1 ]]; then
	    v="${v},"
	fi
	first=0
	v="${v}$1"
	shift
    done
    v="$v]}"
    echo $v
}

# P command pid cpu mem gpu gpumem nproc
P() {
    echo "{\"command\":\"$1\",\"pid\":$2,\"cpu\":$3,\"mem\":$4,\"res\":0,\"gpu\":$5,\"gpumem\":$6,\"nproc\":$7}"
}

output=$($SONALYZE profile -j1119125 -f2023-10-21 --fmt=json,all -- smoketest.csv)
R1=$(T "2023-10-21 08:35" $(P "bwa" 1119426 3075 5 0 0 1))
R2=$(T "2023-10-21 08:40" $(P "bwa" 1119426 3094 5 0 0 1))
R3=$(T "2023-10-21 08:45" $(P "bwa" 1119426 3092 5 0 0 1) $(P "samtools" 1119428 11 259 0 0 1))
R4=$(T "2023-10-21 08:50" $(P "bwa" 1119426 3078 5 0 0 1) $(P "samtools" 1119428 12 259 0 0 1))
R5=$(T "2023-10-21 08:55" $(P "bwa" 1119426 3093 5 0 0 1) $(P "samtools" 1119428 11 259 0 0 1))
R6=$(T "2023-10-21 09:00" $(P "bwa" 1119426 3080 5 0 0 1) $(P "samtools" 1119428 14 262 0 0 1))
R7=$(T "2023-10-21 09:05" $(P "bwa" 1119426 3091 5 0 0 1) $(P "samtools" 1119428 11 262 0 0 1))
R8=$(T "2023-10-21 09:10" $(P "bwa" 1119426 3091 5 0 0 1) $(P "samtools" 1119428 10 262 0 0 1))
R9=$(T "2023-10-21 09:15" $(P "bwa" 1119426 3077 5 0 0 1) $(P "samtools" 1119428 13 262 0 0 1))
expected="[$R1,$R2,$R3,$R4,$R5,$R6,$R7,$R8,$R9]"
CHECK profile_smoketest_json "$expected" "$output"

