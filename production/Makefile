.PHONY: default build fmt generate clean test regress \
	sonar-nodes jobanalyzer-server

default generate fmt:

build clean test regress: sonar-nodes jobanalyzer-server

sonar-nodes jobanalyzer-server:
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
