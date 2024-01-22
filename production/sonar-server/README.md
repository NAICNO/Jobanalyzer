# Sonar server setup

This file normally lives in `$JOBANALYZER/production/sonar-server/README.md`.

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

**IMPORTANT:** Run `build.sh` last, as the test builds may be created with unusual options.

### Step 2: Set up data ingestion and reporting

Create working directories if necessary and copy files, as follows.  The working directory is always
`$HOME/sonar` for whatever user is running the server, if you want something else you need to edit a
bunch of scripts.

```
  cd $JOBANALYZER

  mkdir -p ~/sonar/data ~/sonar/reports ~/sonar/secrets
  chmod go-rwx ~/sonar/secrets

  cp code/infiltrate/infiltrate ~/sonar
  cp code/naicreport/naicreport ~/sonar
  cp code/sonalyze/target/release/sonalyze ~/sonar
  cp code/sonalyzed/sonalyzed ~/sonar

  cp production/sonar-server/cluster-aliases.json ~/sonar
  cp production/sonar-server/POINTER.md ~/sonar
  cp production/sonar-server/server-config ~/sonar
  cp production/sonar-server/*.{sh,cron} ~/sonar
  cp -r production/sonar-server/scripts ~/sonar
```

If `~/sonar/scripts` does not have a subdirectory for your cluster, you will need to create one.  See
"Adding a new cluster" below.

### Step 3: Set up the web server

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

These are my additions to nginx.conf for the default `server`:

```
        # sonalyzed commands
        location /jobs {
                proxy_pass http://localhost:1559;
        }
        location /load {
                proxy_pass http://localhost:1559;
        }
        location /uptime {
                proxy_pass http://localhost:1559;
        }
        location /profile {
                proxy_pass http://localhost:1559;
        }
        location /parse {
                proxy_pass http://localhost:1559;
        }
        location /metadata {
                proxy_pass http://localhost:1559;
        }

        # Dashboard static content
        location / {
                root /data/www;
        }
```

### Step 4: Configure data ingestion and remote query

We must create a password file in `~/sonar/secrets/sonalyzed-auth.txt`.  This is a plaintext file on
`username:password` format, one per line.  It controls access to the query server, `sonalyzed`.

We must create a password file in `~/sonar/secrets/exfil-auth.txt`.  This is a plaintext file on
`username:password` format, one per line.  It controls access to the data infiltration server,
`infiltrate`.

If `infiltrate` is to communicate by HTTPS then we must create HTTPS certificates and keys in
`~/sonar/secrets`.  See [secrets/README.md](secrets/README.md) for more.

Ideally the files in `secrets` are not readable or writable by anyone but the owner, but nobody
checks this.

We must then edit `~/sonar/server-config` to point to the various authorization files, to define the
path to `$DASHBOARD/output`, and to define the ports used for various services.  The ports must be
open for remote access.

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

## Upgrading `infiltrate` and `sonalyzed`

One does not simply copy new executables into place.  Here are some pointers.

`infiltrate` and `sonalyzed` must be spun down on the analysis host by killing them with TERM, once
they are down the executables can be replaced and the start scripts can be run to start new servers.

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
