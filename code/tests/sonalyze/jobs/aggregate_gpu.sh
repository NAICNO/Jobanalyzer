output=$($SONALYZE jobs -u- --min-samples=1 -f 2023-10-04 --fmt=csv,job --no-gpu -- aggregate_gpu.csv)
CHECK no_gpu 1269178 $output

output=$($SONALYZE jobs -u- --min-samples=1 -f 2023-10-04 --fmt=csv,job --some-gpu -- aggregate_gpu.csv)
CHECK some_gpu "" $output

