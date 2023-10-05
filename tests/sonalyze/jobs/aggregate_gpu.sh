#!/bin/bash

export t_name=aggregate_gpu_1
export t_expected=1269178
export t_output=$($SONALYZE jobs -u- --min-samples=1 -f 2023-10-04 -t 2023-10-04 --fmt=csv,job --no-gpu -- aggregate_gpu.csv)
source ../../harness.sh

export t_name=aggregate_gpu_2
export t_expected=""
export t_output=$($SONALYZE jobs -u- --min-samples=1 -f 2023-10-04 -t 2023-10-04 --fmt=csv,job --some-gpu -- aggregate_gpu.csv)
source ../../harness.sh
