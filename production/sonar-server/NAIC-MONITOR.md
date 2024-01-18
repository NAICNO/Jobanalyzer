# naic-monitor.uio.no

Information about the setup of naic-monitor.uio.no.  See bottom for to-do list.

## The system

The original request was for this machine:

- hostname: `naic-monitor.uio.no`
- open to the world
- admin group is `hpc-core`
- admin email is `usit-ai-drift@uio`
- RHEL9, 4 cores, 8GB RAM, 250GB disk
- no cfengine roles
- the tweaks below need to be applied

### Tweak: Include the machine in the hpc_host group

For general setup, the machine needs to have `hpc_host` privileges.  That way, sudo works and login
as root from hrothmund will also work.  Frode had to do this:

```
mreg> policy host_add hpc_host naic-monitor
```

### Tweak: Home directories

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
- "`/etc/sssd/sssd.conf` gets fully managed by cfengine, any change you make there gets reverted again after a few minutes, so you cannot change that file"
- "but you can put modifications to it in `/etc/sssd/conf.d/whatever.conf`"
- "a common mistake (at least for me...) are wrong permissions of files I put there, the permissions need to match `/etc/sssd/sssd.conf`"

**NOTE** home directories then have to be created explicitly for each user under `/home`:
```
mkdir /home/slartibartfast
chown slartibartfast /home/slartibartfast
chgrp slartibartfast /home/slartibartfast
chmod go-rwx /home/slartibartfast
```

### Tweak: Disable selinux enforcement

By default the web server would not serve anything.  This [turned
out](https://stackoverflow.com/questions/25774999/nginx-stat-failed-13-permission-denied) to be
caused by SELinux.  The easiest fix is to just disable enforcement, for now.  That turns out to be
hard because cfengine changes it back.  But there is a way:

- First change `/etc/selinux/config`: set `SELINUX` to `disabled`.
- Then quickly execute `chattr +i /etc/selinux/config` to prevent cfengine from overwriting your changes.
- Then reboot.

### Tweak: Basic web server setup

Installed and configured firewall and nginx as mentioned [here](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/9/html/deploying_web_servers_and_reverse_proxies/setting-up-and-configuring-nginx_deploying-web-servers-and-reverse-proxies).

(The firewall change was surprising given the "open to the world" configuration requested from IT for this VM.)

### Tweak: open ports

If services are to use ports other than the ones that are open by default, they must be opened.
Ports have to be selected carefully, because there is a firewall *outside* naic-monitor and it does
not let everything through.  To see what it lets through, see
[here](https://www-int.usit.uio.no/om/organisasjon/iti/nettdrift/dokumentasjon/nett-info/uio-acl/nexus-xx-gw-2616.acl.txt).
If none of these ports can be made to work, then new ports must be requested.

For the time being, for naic-monitor ingest, query and test I use the open Oracle ports.

You need to
```
# firewall-cmd --permanent --add-ports={nnn/tcp,mmm/tcp}
# firewall-cmd --reload
```
to open ports `nnn` and `mmm` locally.  Again, these need to be ports that are let through by
the external firewall.


### Tweak: Setup disk

The machine came with 250GB raw /dev/sdb.  I initialized this as a physical volume, then added it to
the "internvg" LVM volume group, the apportioned 200GB to /home and another 4GB to /usr, leaving us
about 42GB for whatever we need.

This was very helpful in dealing with LVM: https://tldp.org/HOWTO/LVM-HOWTO/commontask.html, the early
parts of the document also have an intro to LVM.

## Email

With the tweaks in place, email just seems to work, when a test crontab sends mail it ends up in my
standard uio email (webmail).

Per Harold, there *may* be some identity check centrally in email services that the user that sends
the email exists: "I think when sending out emails, the central mail server checks if the sender
address exists - and if it doesn't, it fills it in with `root@ulrik.uio.no` or so".

## Web

The web server is now serving the dashboard.

- Created directory /data/www to hold web data
- Installed nginx as described above
- selinux is disabled as described above
- `/etc/nginx/nginx.conf` is modified as follows: add a route for `/` and then restart the server:
```
        location / {
                root /data/www;
        }
```
- the files `Jobanalyzer/dashboard/*.{html,css,js}` are copied to `/data/www`
- we create a new directory `/var/data/output` to hold the canned, uploaded reports
- i uploaded the current outputs from the NREC vm and they are properly served and everything looks OK.

## STATUS

* Web server is now up.
* Jobanalyzer backend is configured on naic-monitor
* All the normal analysis jobs are running, though of course the deadweight state was not transfered
* `infiltrate` is working.
* ML nodes and Fox are exfiltrating data to this machine,
* `sonalyzed` is working, I can run remote queries (from my machine at home) using curl.
* We have acres of disk.
* Uploads from analysis (ML nodes and Fox) appears to be working.
* ML-nodes moneypenny analysis has been spun down, which means cpuhog / violator reports will not go to RT for the time being, but to me.
* Fox login-1 analysis has been spun down, ditto

## TODO

* The VM *should* have a disk that is persistent, per Sabry, but can we trust this?

Backup is not yet configured:
- Harold says, "for backups we need to contact restore@usit, really just send them a mail and say "we need a backup of [this]""
- mail sent on jan 8

TODO (short term):
- follow up on backup service
- (DONE) preserve all scripts and so on
- (DONE) follow up on disk size
- (DONE) spin down infiltrate on NREC vm
- (DONE) get data over from NREC vm (I have data through jan 8, but I can't put it in place b/c out-of-disk)
- (DONE) spin up sonalyzed on naic-monitor
- (DONE) spin down sonalyzed on NREC vm (can do this soon actually, no new data there)
- (DONE) on jan 10: copy over data from shared disk on ML nodes for jan 9, overwrite whatever's in place
- (DONE) set up analysis cron jobs on naic-monitor, with email to *me*
- check that everything works OK - both email and report generation and report copying
  - (DONE) may need to spin down report copying on moneypenny temporarily?
- (DONE) setup parallel exfil on fox
- (DONE) setup fox analysis on naic-monitor
- when it's all fine, spin down jobs on login-1 PERMANENTLY
- when it's all fine, spin down jobs on moneypenny PERMANENTLY
- when we have disk and backup in place, spin down use of shared disk on ml nodes and start cleaning up

We certainly do not want to keep the selinux disablement hack forever, something else must be put in
its place.  See long discussion with Harold.

We **PROBABLY** want a global user, call it `naic-monitor`, that runs the analysis jobs on
`naic-monitor.uio.no` and performs data ingest and so on?

We **MAY** want to standardize on a system user, call it `naic-daemon`, that runs sonar and other
montor jobs on all the different nodes.  Currently this is `larstha` on the ML nodes, `sonar` on the
Fox compute nodes, but `sonar` is a real user elsewhere so it can't be that on the fox login nodes.

At the same time, complexity: more identities, more complexity.

On the ML nodes, I don't want to run the cron job as me, I really do want a system user to do that.

Down the line, it looks like we can use nginx proxy configurations to support our microservices for
query.  The ports we want to use are generally not open, very annoying.
