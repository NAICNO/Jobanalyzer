SONALYZE=~/go/bin/sonalyze

all: mlx.hpc.uio.no-config.json \
	fox.educloud.no-config.json \
	betzy.sigma2.no-config.json \
	saga.sigma2.no-config.json \
	fram.sigma2.no-config.json
.PHONY: all

mkconf: mkconf.go go.mod
	go build

clean:
	rm -f *-config.json mkconf

# The dependencies aren't quite right here.  Another dependency is whether the sysinfo has changed
# (apart from the timestamp).  I could write code to determine that but I'm not going to.

fox.educloud.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		-cluster fox \
		-name "fox.educloud.no" \
		-aliases "fox" \
		-cross-node \
		-desc "UiO 'Fox' supercomputer" \
	> $@

betzy.sigma2.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		-cluster betzy \
		-name "betzy.sigma2.no" \
		-aliases "betzy" \
		-cross-node \
		-desc "Sigma2 'Betzy' supercomputer" \
	> $@

saga.sigma2.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		-cluster saga \
		-name "saga.sigma2.no" \
		-aliases "saga" \
		-cross-node \
		-desc "Sigma2 'Saga' supercomputer" \
	> $@

fram.sigma2.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		-cluster fram \
		-name "fram.sigma2.no" \
		-aliases "fram" \
		-cross-node \
		-desc "Sigma2 'Fram' supercomputer" \
	> $@

mlx.hpc.uio.no-config.json: Makefile mkconf
	./mkconf \
		-sonalyze $(SONALYZE) \
		-cluster mlx \
		-name "mlx.hpc.uio.no" \
		-aliases "ml,mlx" \
		-desc "UiO 'MLX' quasi-cluster of ML nodes" \
	> $@
