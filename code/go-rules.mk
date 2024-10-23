.PHONY: default build clean test regress

default:

build:
	go build

clean:
	go clean

test:
	go test
	for d in $(SUBDIRS); do (cd $$d ; go test ) ; done
