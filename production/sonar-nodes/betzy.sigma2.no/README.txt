There shall be a directory $SONAR, on a shared disk.  This and the
files therein shall be accessible for read+execute for the user that
will be running sonar, $USER.  For production, $USER should be a
dedicated user, ideally with a system-level UID, though neither is
required.

On many hpc systems, $SONAR is /cluster/shared/sonar, and the scripts
are currently set up to use that.  If that does not match your setup
then edit everything.

The file `scripts.tar` is extracted into $SONAR and the `sonar`
executable is also placed in $SONAR.  (If there is no `scripts.tar`
then generate it with the Makefile.)

The subdirectory $SONAR/secrets/ must be made inaccessible (chown
go-rwx) to everyone but $USER.

The file $SONAR/secrets/upload-auth.netrc must be edited to add the
cluster name and password for the cluster in question.

An executable for `curl` must be available on the system.  The scripts
will use curl to POST data to the upload URL, defined in
sonar-config.sh.  The name of the curl executable can be configured in
sonar-config.sh.

The script sonar-slurm.sh is to be run every five minutes on nodes
that are controlled by slurm.

The script sonar-batchless.sh is to be run every five minutes on nodes
that are not controlled by a batch system (login nodes, interactive
nodes).

To test that your setup works, manually run the latter script from a
login node.  There should be no errors, and an admin with access to
the upload host should be able to see that the data have arrived (can
take up to five minutes due to load balancing).

If you are using cron, look at the sonar-runner-*.cron files.  If not,
you're on your own.

Note, scripts are different on betzy than on other systems as of now
(June 2024) because it has a very old OS.
