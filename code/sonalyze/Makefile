.PHONY: TAGS

TARGET=sonalyze
SUBDIRS=application \
	cmd cmd/cards cmd/clusters cmd/configs cmd/diskprof cmd/gpus cmd/jobs cmd/load \
	cmd/metadata cmd/nodeprof cmd/nodes cmd/parse cmd/profile cmd/report cmd/sacct cmd/snodes \
	cmd/sparts cmd/top cmd/uptime cmd/version \
	common \
	daemon daemon/api0 daemon/api1 daemon/api2 daemon/apiutil \
	daemon/api1/common daemon/api1/cards \
	data/card data/common data/config data/cpusample data/disksample data/gpusample data/node \
	data/nodesample data/sample data/slurmjob data/slurmnode data/slurmpart \
	db db/errs db/filedb db/filesys db/parse db/repr db/special db/types db/util \
	table

include ../go-rules.mk

# Run go install https://github.com/lars-t-hansen/gotags@latest
# You'll also want (setq etags-program-name "gotags") in your .emacs
TAGS:
	find $$(pwd) -name '*.go' | grep -v '/testdata/' | gotags -
