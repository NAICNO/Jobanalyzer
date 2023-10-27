output=$($SONALYZE load --from 2023-10-05 --to 2023-10-05 --fmt=datetime,gpus,host -- gpuset.csv | grep 10:00)
CHECK gpuset_unknown \
      "2023-10-05 10:00  none  ml4
2023-10-05 10:00  unknown  ml5
2023-10-05 10:00  1,3   ml6" \
      "$output"
