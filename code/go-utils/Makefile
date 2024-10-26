SUBDIRS=alias auth config error filesys freecsv gpuset hostglob httpclient \
	httpsrv ini maps options process slices slurm sonalyze status sysinfo time

.PHONY: default build fmt clean test $(SUBDIRS)

default:

build fmt clean test: $(SUBDIRS)

$(SUBDIRS):
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
