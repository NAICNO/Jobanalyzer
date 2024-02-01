# Jobanalyzer server setup

This file normally lives in `$JOBANALYZER/production/jobanalyzer-server/README.md`.

If the VM itself has not been set up then see a later section, "Setting up naic-monitor.uio.no".
The rest of this document assumes that there is a functional Linux server with https web, email,
disk, and backup.

To-do items are generally filed in the issue tracker, not documented here.

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
bunch of scripts.  (The name of the working directory is historical; `$HOME/jobanalyzer` would have
been better.  But it's a pain to change everything.)

```
  cd $JOBANALYZER

  mkdir -p ~/sonar/data ~/sonar/reports ~/sonar/secrets
  chmod go-rwx ~/sonar/secrets

  cp code/infiltrate/infiltrate ~/sonar
  cp code/naicreport/naicreport ~/sonar
  cp code/sonalyze/target/release/sonalyze ~/sonar
  cp code/sonalyzed/sonalyzed ~/sonar

  cp production/jobanalyzer-server/cluster-aliases.json ~/sonar
  cp production/jobanalyzer-server/POINTER.md ~/sonar
  cp production/jobanalyzer-server/server-config ~/sonar
  cp production/jobanalyzer-server/*.{sh,cron} ~/sonar
  cp -r production/jobanalyzer-server/scripts ~/sonar
```

If `~/sonar/scripts` does not have a subdirectory for your cluster, you will need to create one.  See
"Adding a new cluster" below.

### Step 3: Set up the web server

The web server must currently run on the same machine as ingest, analysis, and query.  Generated
reports will be copied from the directory they are generated in and into the web server's data
directory.  The web server also proxies access to the query engine and the ingestion module.

But first, set up the dashboard: The dashboard code must be copied to a suitable directory at or
under the web server's root, we'll call this the dashboard directory, `$DASHBOARD`.  The
configuration used below puts this in `/data/www`.

The directory `$DASHBOARD/output` must exist and must be writable by the user that is going to run the
Jobanalyzer server.

```
  # mkdir -p $DASHBOARD/output
  # chown -R <jobanalyzer-user>:<jobanalyzer-user-group> $DASHBOARD
  # ^D
  $ cd $JOBANALYZER
  $ cp code/dashboard/*.{html,js,css} $DASHBOARD
```

Second, the web server must be configured.  These are my additions to nginx.conf for the default
`server` (see further down about HTTPS setup and so on):

```
        # infiltrate upload points
        location /sonar-reading {
                proxy_pass http://localhost:8086;
        }
        location /sonar-heartbeat {
                proxy_pass http://localhost:8086;
        }

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

The port 1559 is the same used by default for the `sonalyzed` query engine, see next, while 8086
is the port used by default for the `infiltrate` module.

### Step 4: Configure data ingestion and remote query

We must create a password file in `~/sonar/secrets/sonalyzed-auth.txt`.  This is a plaintext file on
`username:password` format, one per line.  It controls access to the query server, `sonalyzed`.

We must create a password file in `~/sonar/secrets/exfil-auth.txt`.  This is a plaintext file on
`username:password` format, one per line.  It controls access to the data infiltration server,
`infiltrate`.

Ideally the files in `secrets` are not readable or writable by anyone but the owner, but nobody
checks this.

We must then edit `~/sonar/server-config` to point to the various authorization files, to define the
path to `$DASHBOARD/output`, and to define the ports used for various services.  The port for
`infiltrate` must be open for remote access.

### Step 5: Activate server

Activate the cron jobs and start the data logger and the query server:

```
  cd ~/sonar
  crontab jobanalyzer.cron
  ./start-infiltrate.sh
  ./start-sonalyzed.sh
```

The data logger and query server run on ports defined in the config file, see above.

## Upgrading `infiltrate` and `sonalyzed`

One does not simply copy new executables into place.

`infiltrate` and `sonalyzed` must be spun down on the analysis host by killing them with TERM, once
they are down the executables can be replaced and the start scripts can be run to start new servers.

## Adding a new cluster

Information about how to set up sonar on the the compute nodes is in [../sonar-nodes/README.md](../sonar-nodes/README.md).

The analysis scripts to run on the Jobanalyzer server are in the subdirectory named for the cluster, eg,
`scripts/mlx.hpc.uio.no`.  These scripts are in turn run by the cron script, [`jobanalyzer.cron`](jobanalyzer.cron).

To add a new cluster, add a new subdirectory in `scripts/` and populate it with appropriate scripts,
probably modifying those from a similar cluster.  Normally you'll want at least scripts to compute
the load reports every 5 minutes, every hour, and every day, and to upload data.  But no scripts are
actually required - cluster data may be available for interactive query only, for example.

In the cluster's script directory there must be a file that describes the nodes in the cluster, its
name must be `CLUSTER-config.json` where `CLUSTER` is the cluster name.  For example,
`mlx.hpc.uio.no-config.json` for the ML nodes cluster.

The process of creating the `CLUSTER-config.json` file has been automated to some extent on systems
that run slurm.  See `../../code/slurminfo`.  It runs `sinfo` and produces a JSON array that is
suitable for the `CLUSTER-config.json` file.  See more documentation in that directory.

The dashboard also needs a few additions in `index.html` and in `code/dashboard/dashboard.js` to
link to the cluster's dashboard and describe the cluster.

See [the PR that added everything for Saga](https://github.com/NAICNO/Jobanalyzer/pull/364) for an
example of what support for a new cluster looks like.

## Setting up naic-monitor.uio.no

Information about the setup of `naic-monitor.uio.no`.

### The system

The original request was for this machine:

- hostname: `naic-monitor.uio.no`
- open to the world
- admin group is `hpc-core`
- admin email is `usit-ai-drift@uio`
- RHEL9, 4 cores, 8GB RAM, 250GB disk
- no cfengine roles

The tweaks below need to be applied.

#### Tweak: Include the machine in the hpc_host group

For general setup, the machine needs to have `hpc_host` privileges.  That way, sudo works and login
as root from hrothmund will also work.  Frode had to do this:

```
mreg> policy host_add hpc_host naic-monitor
```

#### Tweak: Home directories

Home directories are local to the machine (this is a feature) even though user identities are not.
It's probably best for the home directories to be local, it means less confusion about what files
are changed.  But that means there has to be a local setup.

There's an override of home directories in `/etc/sssd/conf.d/homedir_override.conf`.  Per Harold:

```
[root@naic-monitor conf.d]# pwd
/etc/sssd/conf.d
[root@naic-monitor conf.d]# cat homedir_override.conf
[nss]
override_homedir = /home/%u
[root@naic-monitor conf.d]# chmod 600 homedir_override.conf
[root@naic-monitor conf.d]# systemctl restart sssd
[root@naic-monitor conf.d]# getent passwd haroldg
haroldg:*:334263:312914:XXXX:/home/haroldg:/local/gnu/bin/bash
```

Additional info:
- "`/etc/sssd/sssd.conf` gets fully managed by cfengine, any change you make there gets reverted again after
  a few minutes, so you cannot change that file"
- "but you can put modifications to it in `/etc/sssd/conf.d/whatever.conf`"
- "a common mistake (at least for me...) are wrong permissions of files I put there, the permissions need
  to match `/etc/sssd/sssd.conf`"

**NOTE** home directories then have to be created explicitly for each user under `/home`:

```
mkdir /home/slartibartfast
chown slartibartfast /home/slartibartfast
chgrp slartibartfast /home/slartibartfast
chmod go-rwx /home/slartibartfast
```

#### Tweak: Disable selinux enforcement

By default the web server would not serve anything.  This [turned
out](https://stackoverflow.com/questions/25774999/nginx-stat-failed-13-permission-denied) to be
caused by SELinux.  The easiest fix is to just disable enforcement, for now.  That turns out to be
hard because cfengine changes it back.  But there is a way:

- First change `/etc/selinux/config`: set `SELINUX` to `disabled`.
- Then quickly execute `chattr +i /etc/selinux/config` to prevent cfengine from overwriting your changes.
- Then reboot.

#### Tweak: Basic web server setup

Installed and configured firewall and nginx as mentioned
[here](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/deploying_web_servers_and_reverse_proxies/setting-up-and-configuring-nginx_deploying-web-servers-and-reverse-proxies).

(The firewall change was surprising given the "open to the world" configuration requested from IT for this VM.)

#### Tweak: open ports

If services are to use ports other than the ones that are open by default, they must be opened.
Ports have to be selected carefully, because there is a firewall *outside* naic-monitor and it does
not let everything through.  To see what it lets through, see
[here](https://www-int.usit.uio.no/om/organisasjon/iti/nettdrift/dokumentasjon/nett-info/uio-acl/nexus-xx-gw-2616.acl.txt).
If none of these ports can be made to work, then new ports must be requested.

You need to
```
# firewall-cmd --permanent --add-ports={nnn/tcp,mmm/tcp}
# firewall-cmd --reload
```
to open ports `nnn` and `mmm` locally.  Again, these need to be ports that are let through by
the external firewall.

Currently there is no need to open additional ports, as all services proxy via port 443 (HTTPS) with nginx.

#### Tweak: Setup disk

The machine came with 250GB raw /dev/sdb.  I initialized this as a physical volume, then added it to
the "internvg" LVM volume group, the apportioned 200GB to /home and another 4GB to /usr, leaving us
about 42GB for whatever we need.

This was very helpful in dealing with LVM: https://tldp.org/HOWTO/LVM-HOWTO/commontask.html, the early
parts of the document also have an intro to LVM.

### Email

With the tweaks in place, email just seems to work, when a test crontab sends mail it ends up in my
standard uio email (webmail).

Per Harold, there *may* be some identity check centrally in email services that the user that sends
the email exists: "I think when sending out emails, the central mail server checks if the sender
address exists - and if it doesn't, it fills it in with `root@ulrik.uio.no` or so".  Since mail is
sent as me, this has not been an issue.

### Web

The web server serves the dashboard code, pre-created reports, and proxies remote sonalyze queries,
as described earlier.

The HTTPS cert for naic-monitor.uio.no must be obtained from the UiO CA as described
[here](https://www.uio.no/tjenester/it/sikkerhet/sertifikater/kokebok.html).  My `/etc/nginx/nginx.conf`
is modified as follows:

```
    # generated 2024-01-23, Mozilla Guideline v5.7, nginx 1.20.1, OpenSSL 3.0.7, modern configuration, no OCSP
    # https://ssl-config.mozilla.org/#server=nginx&version=1.20.1&config=modern&openssl=3.0.7&ocsp=false&guideline=5.7
    server {
        listen 80 default_server;
        listen [::]:80 default_server;

        location / {
            return 301 https://$host$request_uri;
        }
    }

    server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;

        .... PROXY AND SERVICE STUFF HERE, SEE ABOVE ....

        ssl_certificate /etc/nginx/certificates/naic-monitor.uio.no_fullchain.crt;
        ssl_certificate_key /etc/nginx/certificates/naic-monitor.uio.no.key;
        ssl_session_timeout 1d;
        ssl_session_cache shared:MozSSL:10m;  # about 40000 sessions
        ssl_session_tickets off;

        # modern configuration
        ssl_protocols TLSv1.3;
        ssl_prefer_server_ciphers off;

        # HSTS (ngx_http_headers_module is required) (63072000 seconds)
        add_header Strict-Transport-Security "max-age=63072000" always;
    }
```
