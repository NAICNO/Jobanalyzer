# Make scaffolding for utility libraries.  These are no-ops for build but we do run unit tests.
#
# Free variables:
#
# SUBDIRS is a list of all direct subdirectories with Go code, it can be empty

.PHONY: default build fmt clean test regress

default build:

clean:
	go clean

fmt test:
	go $(MAKECMDGOALS)
	set -e ; for d in $(SUBDIRS); do ( set -e ; cd $$d ; go $(MAKECMDGOALS) ) ; done
