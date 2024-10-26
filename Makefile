SUBDIRS=code production

.PHONY: default build fmt clean test regress $(SUBDIRS)

default:

build fmt clean test regress: $(SUBDIRS)

$(SUBDIRS):
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
