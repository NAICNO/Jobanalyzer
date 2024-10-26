.PHONY: default build fmt clean test

default:

build:

fmt clean:
	go $(MAKECMDGOALS)

test:
	if [[ $$(uname) != Darwin ]]; then go test ; fi
