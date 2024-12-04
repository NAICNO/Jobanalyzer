#!/bin/bash
raw_output=rawdata-4w
cleaned_output=cleaned-4w
computed_output=computed-4w
sheets_name=computed-4w.csv

# Get all OOD jobs
fields=JobName/m30,Duration/sec,CpuAvgPct,CpuPeakPct,RequestedCpus,Cmd
sonalyze jobs -cluster fox -from 4w -fmt awk,noheader,$fields -u - -min-runtime 2h \
    | grep OOD > $raw_output

# Remove some junk jobs
grep -v -E '^OOD ' < $raw_output > $cleaned_output

# Compute utilization columns, add headers, sort data by name
echo $(echo $fields | sed 's/,/ /g') 'ave/alloc% peak/alloc%' > $computed_output
cat $cleaned_output | tail -n +2 \
    | awk '{ print $0, $3/$5, $4/$5 }' \
    | sort \
          >> $computed_output

# Google sheets requires an approved extension to import a file
cp $computed_output $sheets_name
