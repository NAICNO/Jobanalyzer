#!/usr/bin/bash

# Upload generated reports to a web server.

# We need globbing, stay away from -f
set -eu -o pipefail

cluster=betzy.sigma2.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

# It is *important* that we not upload violator/deadweight reports, as they contain PII and are
# served behind authorization only.
upload_files="$report_dir/*-daily.json $report_dir/*-weekly.json $report_dir/*-monthly.json $report_dir/*-quarterly.json"
source $script_dir/upload-subr.sh


