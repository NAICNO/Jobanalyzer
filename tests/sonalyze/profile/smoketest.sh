output=$($SONALYZE profile -j1119125 -f2023-10-21 -- smoketest.csv)
CHECK parse_smoketest \
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
