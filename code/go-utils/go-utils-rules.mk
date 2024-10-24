# Make scaffolding for utility libraries.  These are no-ops for build but we do run unit tests.
#
# Free variables:
#
# SUBDIRS is a list of all direct subdirectories with Go code, it can be empty

.PHONY: default build clean test regress

default build:

clean:
	go clean

test:
	go test
	for d in $(SUBDIRS); do (cd $$d ; go test ) ; done
