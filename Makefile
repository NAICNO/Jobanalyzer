SUBDIRS=code production

.PHONY: default build clean test regress $(SUBDIRS)

default:

build clean test regress: $(SUBDIRS)

$(SUBDIRS):
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
