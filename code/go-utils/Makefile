SUBDIRS=alias auth config error filesys freecsv gpuset hostglob httpclient \
	httpsrv maps options process slices slurm sonalyze status sysinfo time

.PHONY: default build fmt generate clean test $(SUBDIRS)

default:

build fmt generate clean test: $(SUBDIRS)

$(SUBDIRS):
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
