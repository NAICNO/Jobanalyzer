package repr

// This is unusual, since its contents are put together from various sources and it is not a time
// series, though that may change.

type Cluster struct {
	Name        string
	Description string
	Aliases     []string // Not sorted
	ExcludeUser []string // Not sorted
}
