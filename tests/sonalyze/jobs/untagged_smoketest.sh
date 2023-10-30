if [[ $($SONALYZE version) =~ untagged_sonar_data ]]; then
    output=$($SONALYZE jobs -u- --fmt=noheader,host -- untagged_smoketest1.csv untagged_smoketest2.csv | sort | uniq)
    CHECK jobs_untagged_smoketest_hosts "ml8" "$output"

    output=$($SONALYZE jobs -u- --fmt=csv,job,start,end -- untagged_smoketest1.csv untagged_smoketest2.csv | grep -E '^2447150')
    CHECK jobs_untagged_smoketest_times "2447150,2023-06-23 12:25,2023-06-24 09:00" "$output"

    # Translated from a whitebox test, but I'm unsure what this tests, actually...  output, maybe...
    output=$($SONALYZE jobs -u- \
		       --max-cpu-avg=100000000 --max-cpu-peak=100000000 \
		       --max-rcpu-avg=100000000 --max-rcpu-peak=100000000 \
		       --max-gpu-avg=100000000 --max-gpu-peak=100000000 \
		       --max-rgpu-avg=100000000 --max-rgpu-peak=100000000 \
		       -- untagged_smoketest1.csv untagged_smoketest2.csv)
    CHECK jobs_untagged_smoketest_output \
	  "jobm      user      duration  host  cpu-avg  cpu-peak  mem-avg  mem-peak  gpu-avg  gpu-peak  gpumem-avg  gpumem-peak  cmd
4079<     root      1d16h55m  ml8   1        4         1        1         0        0         0           0            tuned
4093!     zabbix    1d17h 0m  ml8   1        5         1        1         0        0         0           0            zabbix_agentd
585616<   larsbent  0d 0h45m  ml8   75       745       194      199       72       84        16          26           python
1649588<  riccarsi  0d 3h20m  ml8   4        140       127      155       38       44        2           2            python
2381069<  einarvid  1d16h55m  ml8   1        2         4        4         0        0         0           0            mongod
1592463   larsbent  0d 2h45m  ml8   28       929       92       116       76       89        20          37           python
1593746   larsbent  0d 2h45m  ml8   74       2498      21       29        52       71        2           3            python
1921146   riccarsi  0d20h50m  ml8   1        97        104      115       38       42        2           2            python
1939269   larsbent  0d 3h 0m  ml8   5        168       116      132       79       92        19          33           python
1940843   larsbent  0d 3h 0m  ml8   6        203       47       62        46       58        2           3            python
2126454   riccarsi  0d 6h45m  ml8   2        99        149      149       57       59        2           3            python
2447150   larsbent  0d20h35m  ml8   1        173       18       19        0        0         1           1            python
2628112   riccarsi  0d14h 0m  ml8   1        91        147      148       57       61        2           3            python
2640656   larsbent  0d 1h25m  ml8   82       1462      102      104       64       93        19          38           python
2643165   larsbent  0d 1h25m  ml8   9        152       37       41        60       86        3           3            python
2722769   larsbent  0d11h20m  ml8   8        1071      121      140       79       93        22          40           python,python <defunct>
2722782>  larsbent  0d11h25m  ml8   2        170       61       88        55       84        2           3            python
2727498   adamjak   0d 2h45m  ml8   1        21        1        1         0        0         0           0            node
2747449   adamjak   0d 0h20m  ml8   5        22        1        1         0        0         0           0            python
2750031   adamjak   0d 0h15m  ml8   25       100       1        1         0        0         0           0            python" \
	  "$output"
fi
