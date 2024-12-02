.PHONY: default build fmt generate clean test

default:

build:

fmt generate clean:
	go $(MAKECMDGOALS)

test:
	if [[ $$(uname) != Darwin ]]; then go test ; fi
