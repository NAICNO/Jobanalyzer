package sonalyze

// Constants defined jointly with ~/sonalyze/src/jobs.rs: these are bits in the output from the
// "classification" fields of a jobs report.

const (
	LIVE_AT_END   = 1 // Latest timestamp coincides with last record read
	LIVE_AT_START = 2 // Ditto earliest/first
)
