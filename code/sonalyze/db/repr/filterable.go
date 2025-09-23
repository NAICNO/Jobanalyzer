package repr

// `Filterable` is implemented by data representations that can be filtered by the standard
// ApplyFilter function in data/common/filter.go.  The timeVal must be either an RFC3339 timestamp
// string or a time.Time value.  If a string, it will be parsed into a time value.

type Filterable interface {
	TimeAndNode() (timeVal any, nodeName string)
}
