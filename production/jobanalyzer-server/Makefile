# Rebuild the cron scripts from sources in cronstubs/

.SUFFIXES: .cron .stub
.PHONY: all

all: naicreport.cron sonalyzed.cron sonalyzed-and-naicreport.cron

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
