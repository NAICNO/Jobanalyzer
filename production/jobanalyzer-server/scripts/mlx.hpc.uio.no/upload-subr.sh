# This is shared code, to be sourced from other scripts.  It uploads/copies the generated reports to
# the web server's directory.
#
# If the variable NOUPLOAD is set in the environment then no uploading is done, but the file names
# to be uploaded are echoed instead.
#
# Free variables:
#
#   upload_files - expanded list of file names to upload
#   naicreport_upload_dir - the directory into which to copy those files

# The chmod is done here so that we don't have to do it in the generating scripts or on the server,
# and we won't have to depend on the umask.  But it must be done, or the files may not be readable
# by the web server.

chmod go+r $upload_files

if [[ -v NOUPLOAD ]]; then
    echo $upload_files
else
    cp $upload_files $naicreport_upload_dir
fi
