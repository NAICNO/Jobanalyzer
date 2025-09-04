# These are settings for running as user sonalyze on naic-monitor

SONALYZE=$(HOME)/sonalyze/sonalyze
CREDENTIAL=-auth-file $(HOME)/secrets/naicreport/credential.txt -remote https://naic-monitor.uio.no

default:
	@echo "Adjust the variables in the Makefile and run 'make clean ; make config'"
.PHONY: default

config: mlx.hpc.uio.no-config.json \
	fox.educloud.no-config.json \
	betzy.sigma2.no-config.json \
	saga.sigma2.no-config.json \
	fram.sigma2.no-config.json
.PHONY: config

build: mkconf
.PHONY: build

clean:
	go clean
	rm -f *-config.json
.PHONY: clean

mkconf: mkconf.go go.mod ../go-utils/*/*.go
	go build

fmt generate test:
	go $(MAKECMDGOALS)
	set -e ; for d in $(SUBDIRS); do ( set -e ; cd $$d ; go $(MAKECMDGOALS) ) ; done
.PHONY: fmt generate test

# The dependencies aren't quite right here.  Another dependency is whether the sysinfo has changed
# (apart from the timestamp).  I could write code to determine that but I'm not going to.

fox.educloud.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		$(CREDENTIAL) \
		-cluster fox \
		-name "fox.educloud.no" \
		-aliases "fox" \
		-cross-node \
		-desc "UiO 'Fox' supercomputer" \
	> $@

betzy.sigma2.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		$(CREDENTIAL) \
		-cluster betzy \
		-name "betzy.sigma2.no" \
		-aliases "betzy" \
		-cross-node \
		-desc "Sigma2 'Betzy' supercomputer" \
	> $@

saga.sigma2.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		$(CREDENTIAL) \
		-cluster saga \
		-name "saga.sigma2.no" \
		-aliases "saga" \
		-cross-node \
		-desc "Sigma2 'Saga' supercomputer" \
	> $@

fram.sigma2.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		$(CREDENTIAL) \
		-cluster fram \
		-name "fram.sigma2.no" \
		-aliases "fram" \
		-cross-node \
		-desc "Sigma2 'Fram' supercomputer" \
	> $@

mlx.hpc.uio.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		$(CREDENTIAL) \
		-cluster mlx \
		-name "mlx.hpc.uio.no" \
		-aliases "ml,mlx" \
		-desc "UiO 'MLX' quasi-cluster of ML nodes" \
	> $@
