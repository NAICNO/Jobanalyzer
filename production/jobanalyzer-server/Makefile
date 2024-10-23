# Rebuild the cron scripts from sources in cronstubs/

.SUFFIXES: .cron .stub
.PHONY: default build clean test regress

default:

build: naicreport.cron sonalyzed.cron sonalyzed-and-naicreport.cron

# Do nothing for clean, I think, since the .cron files are checked in to the repo.
# We could do git restore *.cron but I'm not sure that's quite right for the buildbot.
clean:

test:

regress:
	$(MAKE) build
	for f in cluster-config/*-config.json cluster-config/cluster-aliases.json ; do \
		echo $$f ; \
		../../code/jsoncheck/jsoncheck $$f ; \
	done

naicreport.cron: \
	cronstubs/do-not-edit.stub \
	cronstubs/naicreport-header.stub \
	cronstubs/mailto.stub \
	cronstubs/naicreport.stub
	cat $^ > $@

sonalyzed.cron: \
	cronstubs/do-not-edit.stub \
	cronstubs/sonalyzed-header.stub \
	cronstubs/mailto.stub \
	cronstubs/sonalyzed.stub
	cat $^ > $@

sonalyzed-and-naicreport.cron: \
	cronstubs/do-not-edit.stub \
	cronstubs/sonalyzed-and-naicreport-header.stub \
	cronstubs/mailto.stub \
	cronstubs/sonalyzed.stub \
	cronstubs/naicreport.stub
	cat $^ > $@
