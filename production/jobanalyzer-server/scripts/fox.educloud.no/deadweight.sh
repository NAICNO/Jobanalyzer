#!/usr/bin/bash
#
# Run sonalyze for the `deadweight` use case and capture its output in a file appropriate for the
# current time and system.

set -euf -o pipefail

cluster=fox.educloud.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

output_dir=${state_dir}/$(date +'%Y/%m/%d')
mkdir -p ${output_dir}

# Report jobs that are not doing anything useful but are hanging onto system resources (zombies,
# defuncts, and maybe others).  This is defined entirely by the sonalyze `--zombie` option, for now.
#
# The report is run on the data for the last 24h.  This should therefore be run about once every 12h
# at least, but ==> IMPORTANTLY, it MUST be run often enough that job IDs are not reused between
# consecutive runs.  It is not expensive, and can be run fairly often.

$naicreport_dir/sonalyze jobs \
		    $data_source \
		    -u - \
		    --zombie \
		    --fmt=csvnamed,tag:deadweight,now,std,start,end,cmd \
		    "$@" \
		    >> ${output_dir}/deadweight.csv
