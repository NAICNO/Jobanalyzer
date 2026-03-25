# Running the sonalyze daemon

In the "classical" daemon mode, sonalyze manages a simple time series database in a directory tree
and presents a REST API (actually two APIs) for querying the data.  One of the APIs can be used to
insert data as well, or sonalyze can be told to connect to a Kafka broker to listen for new data.

In the "new" daemon mode, sonalyze can alternatively be run on a timescaledb database.  In this
case, sonalyze can only serve queries - database creation and data insertion must be performed by
other means, normally by [slurm-monitor](https://github.com/2maz/slurm-monitor), on whose database
schema sonalyze is dependent.

