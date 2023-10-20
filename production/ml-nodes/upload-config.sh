# This file must define the following variables.  The variable $sonar_dir will name the root of the
# sonar work directory.

# A file to be used as the argument to scp's "-i" argument
IDENTITY_FILE_NAME=$sonar_dir/ubuntu-vm.pem

# The user and host to receive uploaded data
WWWUSER_AND_HOST=ubuntu@158.39.48.160

# The directory within that host in which to place data
WWWUSER_UPLOAD_PATH=/var/www/html/output
