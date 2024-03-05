package time

const (
	// This format is used by some outputs, while RFC3339 is used by other outputs - it's generally
	// confusing, plus there are conflicting needs for human-readable vs machine-readable text.  See
	// comments on https://github.com/NAICNO/Jobanalyzer/issues/436 for more information.  Searching
	// for either `RFC3339` or `CommonDateTimeFormat` in Go code should find all date formatting
	// code.  In Rust code, probably searching for `%Y` is the best bet.
	CommonDateTimeFormat = "2006-01-02 15:04"
)
