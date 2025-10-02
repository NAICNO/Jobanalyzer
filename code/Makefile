# `build` builds all the programs
#
# `clean` and `test` cleans programs and libraries
#
# `regress` builds programs, then runs the regression test shell scripts in various directories,
# which must be added below.  Not every directory has regression tests.

SUBDIRS=sonalyze naicreport sonard slurminfo make-cluster-config jsoncheck netsink numdiff \
	heatmap generate-table go-utils tests dashboard

.PHONY: default build fmt generate clean test regress $(SUBDIRS)

default:

build fmt generate clean test: $(SUBDIRS)

regress: regress-build tests

regress-build:
	$(MAKE) build

$(SUBDIRS):
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
