# Make scaffolding for programs.
#
# Free variables:
#
# TARGET is the name of the target being built, for `build`
# SUBDIRS is a list of all direct subdirectories with Go code, it can be empty

.PHONY: build fmt vet generate clean test regress

$(TARGET): go.mod *.go ../go-utils/*/*.go $(SUBDIRS:=/*.go)
	go build

build: $(TARGET)

clean:
	go clean

fmt generate test vet:
	go $(MAKECMDGOALS)
	set -e ; for d in $(SUBDIRS); do ( set -e ; cd $$d ; go $(MAKECMDGOALS) ) ; done
