.PHONY: TAGS

TARGET=sonalyze
SUBDIRS=application \
	cmd cmd/add cmd/cards cmd/clusters cmd/configs cmd/gpus cmd/jobs cmd/load cmd/metadata \
	cmd/nodeprof cmd/nodes cmd/parse cmd/profile cmd/report cmd/sacct cmd/snodes cmd/sparts \
	cmd/top cmd/uptime cmd/version \
	common \
	daemon \
	data/card data/common data/config data/cpusample data/gpusample data/node data/sample \
	data/slurmjob data/slurmnode data/slurmpart \
	db db/errs db/filedb db/parse db/repr db/special db/types \
	table

include ../go-rules.mk

# Run go install https://github.com/lars-t-hansen/gotags@latest
# You'll also want (setq etags-program-name "gotags") in your .emacs
TAGS:
	find $$(pwd) -name '*.go' | grep -v '/testdata/' | gotags -
