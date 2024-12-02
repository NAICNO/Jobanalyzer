SUBDIRS=code production

.PHONY: default build fmt generate clean test regress $(SUBDIRS)

default:

build fmt generate clean test regress: $(SUBDIRS)

$(SUBDIRS):
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
