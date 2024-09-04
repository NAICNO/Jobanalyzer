#!/bin/bash
#
# Before running this, go to ../../code and run build.sh
#
# This is currently meant to run on naic-report with direct access to the data stores (multiple
# stores, one for production and one for experimental slurm data).  But it should be easy enough to
# change this to remote access to a single store later.
#
# Note for png output, pnmtopng must be installed, usually via netpbm and/or netpbm-apps or somesuch
#
# We exclude gpu-3 because it is dedicated to interactive work and has no Slurm jobs.

source config.bash
GPUS='gpu-[1,2,4-13]'

# Excluding the interactive node, map all jobs that ran for at least 2h
# $SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
#           -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPUS" -min-runtime 2h \
#           -fmt awk,job,sgpu,sgpumem \
#   | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-map.txt

# Same, but as png
# $SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
#           -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPUS" -min-runtime 2h \
#           -fmt awk,job,sgpu,sgpumem \
#     | $HEATMAP -a 2 -b 4 -n 10 -ppm \
#     | pnmtopng > serious-foxgpu-map.png

# And also exclude interactive-ish jobs.  This means combining sonar and slurm data.  The way we do
# this is extract all interactive job numbers from slurm, then exclude those when running sonalyze.
interactive=$($SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY \
                              -all \
                              -fmt awk,JobID,User,Account,State,JobName \
                  | grep -E 'interactive|OOD|ood|TIMEOUT|CANCELLED' \
                  | awk '{ if (s != "") { s = s "," }; s = s $1 } END { print s }')

# All GPUs

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPUS" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-noninteractive-map.txt

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPUS" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 -ppm \
    | pnmtopng > serious-foxgpu-noninteractive-map.png

# GPU classes.  Skip gpu-3 for same reasons as above
GPU_A100_40='gpu-[1,2,13]'
GPU_RTX3090_24='gpu-[4,5,6,11]'
GPU_A100_80='gpu-[7,8]'
GPU_H100_80='gpu-[9,10]'
GPU_A40_44='gpu-12'

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPU_A100_40" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-a100-40-noninteractive-map.txt

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPU_RTX3090_24" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-rtx3090-24-noninteractive-map.txt

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPU_A100_80" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-a100-80-noninteractive-map.txt

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPU_H100_80" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-h100-80-noninteractive-map.txt

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPU_A40_44" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem \
    | $HEATMAP -a 2 -b 4 -n 10 > serious-foxgpu-a40-44-noninteractive-map.txt
