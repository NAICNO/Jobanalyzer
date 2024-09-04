#!/bin/bash
#
# Before running this, go to ../../code and run build.sh
#
# This is currently meant to run on naic-report with direct access to the data stores (multiple
# stores, one for production and one for experimental slurm data).  But it should be easy enough to
# change this to remote access to a single store later.
#
# We exclude gpu-3 because it is dedicated to interactive work and has no Slurm jobs.
#
# Maybe gpu-9 also??

source config.bash
GPUS='gpu-[1,2,4-13]'

# And also exclude interactive-ish jobs.  This means combining sonar and slurm data.  The way we do
# this is extract all interactive job numbers from slurm, then exclude those when running sonalyze.
interactive=$($SONALYZE_SACCT -from $FIRSTDAY -to $LASTDAY \
                              -all \
                              -fmt awk,JobID,User,Account,State,JobName \
                  | grep -E 'interactive|OOD|ood|TIMEOUT|CANCELLED' \
                  | awk '{ if (s != "") { s = s "," }; s = s $1 } END { print s }')

$SONALYZE jobs -data-dir $SONARDATADIR -config-file $CONFIGFILE \
          -some-gpu -from $FIRSTDAY -to $LASTDAY -u - -host "$GPUS" -min-runtime 2h \
          -exclude-job ${interactive} \
          -fmt awk,job,sgpu,sgpumem,user,cmd | awk '
{ if ($3 <= 25 && $5 <= 25) {
    print $7
  }
}
'



