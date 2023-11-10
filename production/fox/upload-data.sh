#!/usr/bin/bash

# Upload generated reports to a web server.
#
# NOTE!  If upload logic needs to change here, also consider load-5m-and-upload.sh, which performs
# its own upload.

# We need globbing, stay away from -f
set -eu -o pipefail

sonar_dir=$HOME/sonar
sonar_bin=$sonar_dir
data_dir=$sonar_dir/data
output_dir=$sonar_dir/output

# The chmod is done here so that we don't have to do it in naicreport or on the server,
# and we don't depend on the umask.  But it must be done, or the files may not be
# readable by the web server.
chmod go+r $output_dir/*.json

source $sonar_dir/upload-config.sh

# StrictHostKeyChecking has to be disabled here because this is not an interactive script,
# and the VM has not been configured to respond in such a way that the value in known_hosts
# will bypass the interactive prompt.
scp -C -q -o StrictHostKeyChecking=no -i $IDENTITY_FILE_NAME \
    $output_dir/*.json \
    $WWWUSER_AND_HOST:$WWWUSER_UPLOAD_PATH


