.PHONY: default build fmt clean test regress \
	sonar-nodes jobanalyzer-server

default fmt:

build clean test regress: sonar-nodes jobanalyzer-server

sonar-nodes jobanalyzer-server:
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
