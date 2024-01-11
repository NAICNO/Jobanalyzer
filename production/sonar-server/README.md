# Sonar server setup

This file normally lives in `$JOBANALYZER/production/sonar-server/README.md`.

FIXME: The following instructions are pretty rough and they get rougher toward the bottom.

## Fundamentals

Let `$JOBANALYZER` be the Jobanalyzer source code root directory.

### Step 1: Build

First build and test the executables, as follows.  Remember, you may first need to `module load` Go
1.20 or later and Rust 1.65 or later, or otherwise obtain those tools.

```
  cd $JOBANALYZER/code
  ./run_tests.sh
  ./build.sh
```

**IMPORTANT:** Run `build.sh` last as the test builds may be created with unusual options.

### Step 2: Setup data ingestion and reporting

Create working directories if necessary and copy files, as follows.  The working directory is always
`$HOME/sonar` for whatever user is running the server, if you want something else you need to edit a
bunch of scripts.

```
  cd $JOBANALYZER

  mkdir -p ~/sonar/data ~/sonar/reports ~/sonar/q

  cp code/infiltrate/infiltrate ~/sonar
  cp code/sonalyzed/sonalyzed ~/sonar
  cp code/naicreport/naicreport ~/sonar
  cp code/sonalyze/target/release/sonalyze ~/sonar

  cp code/sonalyzed/query/*.{html,js,css} ~/sonar/q

  cp production/sonar-server/server-config ~/sonar
  cp production/sonar-server/cluster-aliases.json ~/sonar
  cp production/sonar-server/POINTER.md ~/sonar
  cp production/sonar-server/*.{sh,cron} ~/sonar
  cp -r production/sonar-server/scripts ~/sonar
```

If `~/sonar/scripts` does not have a subdirectory for your cluster, you will need to create one.  See
"Adding a new cluster" below.

### Step 3: Setup the web server

The web server must currently run on the same machine as ingest and analysis.  The reports will be
copied from the directory they are generated in and into the web server's data directory.

But first, set up the dashboard: The dashboard code must be copied to a suitable directory at or
under the web server's root, we'll call this the dashboard directory, `$DASHBOARD`.
```
  cd $JOBANALYZER
  cp code/dashboard/*.{html,js,css} $DASHBOARD
```

The directory $DASHBOARD/output must exist and must be writable by the user that is going to run the
sonar server, eg:
```
  # mkdir -p $DASHBOARD/output
  # chown -R <sonar-user>.<sonar-user-group> $DASHBOARD/output
```

### Step 4: Configure

We must set up a password file, usually in `~/.ssh/sonalyzed-auth.txt`.  This is a plaintext file on
`username:password` format, one per line.  It controls access to the query server, `sonalyzed`.

We must also set up an identity file, usually in `~/.ssh/exfil-auth.txt`.  This is a plaintext file
with a single line on `username:password` format.  It is used to authorize data sent to the
infiltration server, `infiltrate`.  It is the same file that `exfiltrate` uses on the compute node
to identify itself to the server.

Ideally these files are not readable or writable by anyone but the owner, but nobody checks this.

We must then edit `~/sonar/server-config` to point to the authorization files, to define the path to
`$DASHBOARD/output`, and to define the ports used for various services.  The ports must be open for
remote access.

We must finally edit `$DASHBOARD/testflag.js` (which is the configuration file for the dashboard) to
set up at least the variable `SONALYZED` to point at the correct host and port for the sonalyze
daemon.

### Step 5: Activate server

Activate the cron jobs and start the data logger and the query server:

```
  cd ~/sonar
  crontab sonar-server.cron
  ./start-infiltrate.sh
  ./start-sonalyzed.sh
```

The data logger and query server run on ports defined in the config file, see above.

## Upgrading `infiltrate` and `exfiltrate`

One does not simply copy new executables into place.  Here are some pointers.

Infiltrate must be spun down on the analysis host by killing it with HUP, once it is down the executable
can be replaced and `start-infiltrate.sh` can be run again to start a new server.

The `exfiltrate` executable is used by the golang runtime on all the cluster nodes and overwriting
it with new data will cause the runtime to crash.  Also, given the random sending window there is no
one time when the executable is not busy.  Probably (untested), it is enough to mv the new
executable into the place of the old one, so as to replace it atomically, and hope that the runtime
has gone through the proper channels to hold onto a reference to its image from where it's linked in
`/proc/pid`.

## Data logger daemon

### Basics

To start the infiltration server (data logging daemon), we need to run this in the background:

```
./infiltrate -data-path ~/sonar/data -port 8086 -auth-file ~/.ssh/exfil-auth.txt
```

This will listen for incoming data on HTTP (not HTTPS!) from all the nodes in all the clusters and
will log data in Sonar-format files in `~/sonar/data/<cluster>/<year>/<month>/<day>/<hostname>.csv`.

The file `~/.ssh/exfil-auth.txt` contains one line of text on the form `<username>:<password>`,
characters should be printable ASCII and `<username>` cannot contain a `:`.

The sending side must use the same credentials and port.

### Startup on boot, and remaining up

There is a crontab, sonar-server.cron, that has a @reboot spec for this program, and it runs the
file start-infiltrate.sh.

## Adding a new cluster

The analysis scripts to run on the server are in the subdirectory named for the cluster, eg,
`scripts/mlx.hpc.uio.no`.  These scripts are in turn run by the cron script, `sonar-server.cron`.

To add a new cluster, add a new subdirectory and populate it with appropriate scripts.  Normally
you'll want at least scripts to compute the load reports every 5 minutes, every hour, and every day,
and to upload data.  But no scripts are actually required - cluster data may be available for
interactive query only, for example.

In the cluster's script directory there must be a file that describes the nodes in the cluster, its
name must be `CLUSTER-config.json` where `CLUSTER` is the cluster name.  For example,
`mlx.hpc.uio.no-config.json` for the ML nodes cluster.

## Bugs, future directions

## Data logger (infiltrate)

We need to support HTTPS/TLS.

In the future it would be useful for the password file to be able to contain multiple lines so that
different clusters could have their own credentials.  Sonalyzed already allows this.

We need something so that this server is "always" up without the user crontab.  There are a couple
of ways of doing this.  One is to integrate with systemd, which is probably right in the long run.
The other is to run a script every 15 minutes or so (in any case well within the window that the
transport can deal with) that checks the infiltrate.pid file and performs a pgrep infiltrate, and if
things don't check out (the values obtained have to be equal) then (re)starts the server if
necessary.

There should be a "ping" function in the server, ie an address "/ping" that will return 200 OK to
show that the server is up.  It makes it possible to check not just that the process is there but
that it is responsive.

## Interactive query (sonalyzed)

We need to support HTTPS/TLS.

We need something to ensure that the server remains up / is restarted if it crashes.
