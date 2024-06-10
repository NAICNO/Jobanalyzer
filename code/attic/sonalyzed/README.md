# Sonalyzed

`sonalyzed` is an HTTP server that runs sonalyze on behalf of a remote client.  It responds to `GET`
and POST requests carrying parameters that specify how to run sonalyze against a local data store
and how to insert data into the store.

See `sonalyzed.go` for information about how to configure and run the server.

