#!/usr/bin/bash

# Upload generated reports to a web server.

# We need globbing, stay away from -f
set -eu -o pipefail

cluster=fram.sigma2.no
naicreport_dir=${naicreport_dir:-$HOME/sonar}
source $naicreport_dir/naicreport-config

upload_files="$report_dir/*.json"
source $script_dir/upload-subr.sh


