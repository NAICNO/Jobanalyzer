output=$($SONALYZE metadata --from 2023-09-01 -- metadata_smoketest.csv)
CHECK metadata_smoketest \
      "ml1.hpc.uio.no,2023-09-02 22:00,2023-09-10 22:05
ml8.hpc.uio.no,2023-09-01 22:00,2023-09-13 22:00" \
      "$output"
