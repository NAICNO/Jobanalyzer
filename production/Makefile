.PHONY: default build clean test regress \
	sonar-nodes jobanalyzer-server

default:

build clean test regress: sonar-nodes jobanalyzer-server

sonar-nodes jobanalyzer-server:
	( cd $@ ; $(MAKE) $(MAKECMDGOALS) )
