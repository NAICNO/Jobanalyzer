# Remote access to sonar data

You will receive a user name / passwd for the database via some channel.  You can use these
credentials to identify yourself in two situations: for all remote command-line access, and when
you're being asked for them in the dashboard on https://naic-monitor.uio.no in a web browser (some
data are public there, only job queries are behind a passwd).  This GUI is what most people think of
as Jobanalyzer, but the command line interface is much richer and is really the primary access path
for all serious use.

Place your user name and password in a file on the form you received them, I usually put them in
`~/.ssh/sonalyzed-auth.netrc` and make sure it's properly protected.

Do ask questions.


## Source code repo, filing bugs, getting help.

Jobanalyzer repo: https://github.com/NAICNO/Jobanalyzer.git.

I expect you will find bugs or will want features we don't have. If you see something, file
something: https://github.com/NAICNO/Jobanalyzer/issues.  Please be liberal about filing issues -
better too many than too few.

There is a manual: `code/sonalyze/MANUAL.md`.  It's a little rudimentary but we try to keep it up to
date and it is better than the on-line help.

When you run the `sonalyze` binary (see below) tou can try `sonalyze help`, but this is mostly for
people who know what's going on already and just need a reminder of what the name of an option is
and things like that.

The output format for many operations is controlled with the `-fmt` switch to sonalyze.  Use `-fmt
help` to get an overview of the options for the command in question.


## Command line access

Clone the repo, then `make build`at the top level should be sufficient.  (You need Go 1.22.1 or
newer, this is not a problem on most systems, certainly not most personal systems.)  You now want to
run `code/sonalyze/sonalyze`.

The program `sonalyze` just constructs an HTTPS query and executes it (and I use it because it's
convenient and does a fair bit of error checking for me); you could accomplish the same directly via
`curl` or from any other kind of program that can do HTTPS, the protocol is documented in
`code/sonalyze/REST.md`.

The template for a command is

```
sonalyze <command> -remote https://naic-monitor.uio.no -auth-file ~/.ssh/sonalyzed-auth.netrc -cluster <clustername> ...
```

You can test that your setup is working by executing this:

```
sonalyze cluster -remote https://naic-monitor.uio.no -auth-file ~/.ssh/sonalyzed-auth.netrc
```

You'll see a list of clusters known to Jobanalyzer.

Tip: You can place the values for -remote and -auth-file in a ~/.sonalyze file so that you don't have to
type them every time, see the manual.


## Example

Suppose you want to execute a query every few minutes with the purpose of generating an alert if a
host is down.  I might do this every 5-10 minutes:

```
sonalyze meta --remote  https://naic-monitor.uio.no -auth-file ~/.ssh/sonalyzed-auth.txt -cluster fox -bounds -from 2w -fmt host,latest
```

which says we want to examine hosts on the `fox` cluster that have been alive during the last two
weeks (`-from 2w`), extract the bounds of the data stream (earliest/latest) and print the host name
and the latest timestamp from each.  In this case the output is CSV (but JSON and other formats are
possible).  The output looks like this (edited):

```
bigmem-2.fox,2024-10-30 06:55
c1-10.fox,2024-10-30 07:00
c1-11.fox,2024-10-30 07:00
...
gpu-6.fox,2024-10-30 07:00
gpu-7.fox,2024-10-24 08:25
gpu-8.fox,2024-10-30 06:55
gpu-9.fox,2024-10-30 07:00
int-1.fox.ad.fp.educloud.no,2024-10-30 06:55
...
login-4.fox.ad.fp.educloud.no,2024-10-30 06:55
```

and by looking at these data it's clear that gpu-7 has been down for 6 days.  Time stamps vary a
little because data arrive asynchronously but if a timestamp is 15 minutes old or more you can
assume the host has stopped reporting.

(There are other ways of doing the same thing and some of them might be better, but the above has
the benefit of being simple.)

There are more example queries in `~/adhoc-reports/`.  A simple one is `top-commands.sh`, which
shows top compute consumers on Saga during the last week.  A more complex one is
`~/adhoc-reports/leadership-report/bad-gpu-jobs.py`, which also uses Slurm data and this currently
makes it specific to Fox since we're not extracting Slurm data from Saga/Fram/Betzy yet.
