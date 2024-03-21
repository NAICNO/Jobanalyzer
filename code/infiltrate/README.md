# Infiltrate

`infiltrate` - data receiver for Sonar data, to run on database/analysis host.

## How to run

See comment block at the beginning of infiltrate.go for all information.

## Staying up

The infiltration agent needs to always be running.  Some system external to the agent (eg systemd)
needs to ensure that it is.

## Future developments

Currently, infiltrate receives CSV and JSON data by HTTP POST and stores them as in the existing
sonarlog directory tree in the format received (sonar readings are CSV, sysinfo data are JSON).

In the future we may move to another protocol and architecture, eg, mqtt-based transport, where
infiltrate subscribes to messages from the broker.  We may also use a proper database.
