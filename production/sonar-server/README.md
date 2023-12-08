# Sonar server setup

This file normally lives in $JOBANALYZER/production/sonar-server/README.md.

## Fundamentals

Let $JOBANALYZER be the Jobanalyzer root directory.

First build and test the executables:

```
  cd $JOBANALYZER
  ./run_tests.sh
  ./build.sh
```

You may need to install tools if this fails along the way.  It's important to run `build.sh` last as
the test builds may be created with unusual options.

Create working directories if necessary and copy files.  The working directory is always
`$HOME/sonar`, if you want something else you need to edit a bunch of scripts.

```
  cd $JOBANALYZER
  mkdir -p ~/sonar/data
  cp infiltrate/infiltrate ~/sonar
  cp production/sonar-server/start-infiltrate.sh ~/sonar
  cp production/sonar-server/sonar-server.cron ~/sonar
  cp production/sonar-server/README.md ~/sonar
```

Set up the crontab and start the data logger:

```
  cd ~/sonar
  crontab sonar-server.cron
  ./start-infiltrate.sh
```

## Upgrading infiltrate and exfiltrate

One does not simply copy new executables into place.  Here are some pointers.

Infiltrate must be spun down on the analysis host by killing it with HUP, once it is down the executable
can be replaced and start-infiltrate.sh can be run again to start a new server.

The exfiltrate executable is used by the golang runtime on all the cluster nodes and overwriting it
with new data will cause the runtime to crash.  Also, given the random sending window there is no
one time when the executable is not busy.  Probably (untested), it is enough to mv the new
executable into the place of the old one, so as to replace it atomically, and hope that the runtime
has gone through the proper channels to hold onto a reference to its image from where it's linked in
/proc/pid.

## Data logger daemon

### Basics

To start the infiltration server (data logging daemon), we need to run this in the background:

```
./infiltrate -data-path ~/sonar/data -port 8086 -auth-file ~/.ssh/exfil-auth.txt
```

This will listen for incoming data on HTTP (not HTTPS!) from all the nodes in all the clusters and
will log data in Sonar-format files in `~/sonar/data/<cluster>/<year>/<month>/<day>/<hostname>.csv`.

The file ~/.ssh/exfil-auth.txt contains one line of text on the form `<username>/<password>`,
characters should be printable ASCII and `<username>` cannot contain a `:`.

The sending side must use the same credentials and port.

### Startup on boot, and remaining up

There is a crontab, sonar-server.cron, that has an @boot spec for this program, and it runs the file
start-infiltrate.sh.

### Bugs, future directions

We need to support HTTPS/TLS.

In the future it would be useful for the password file to be able to contain multiple lines so that
different clusters could have their own credentials.

We need something so that this server is "always" up without the user crontab.  There are a couple
of ways of doing this.  One is to integrate with systemd, which is probably right in the long run.
The other is to run a script every 15 minutes or so (in any case well within the window that the
transport can deal with) that checks the infiltrate_pid file and performs a pgrep infiltrate, and if
things don't check out (the values obtained have to be equal) then (re)starts the server if
necessary.

There should be something that terminates the server properly on reboot, ie sends it HUP.  Not sure
what the OS will do.

There should be a "ping" function in the server, ie an address "/ping" that will return 200 OK to
show that the server is up.  It makes it possible to check not just that the process is there but
that it is responsive.

## Analyses

The analysis scripts to run on the server are in the directory named for the cluster, eg, mlx.hpc.uio.no
