.PHONY: default build clean test

default:

build:
	go build

clean:
	go clean

test:
	if [[ $$(uname) != Darwin ]]; then go test ; fi
