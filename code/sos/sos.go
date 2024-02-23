package main

import (
	"path"
	"regexp"
	"time"
)

type eventBlob struct {
	dir path.Path
	hostData map[string][]*EventRecord
}

type Store struct {
	rootDir path.Path

	// The key here is cluster/year/month/day/hostname where all times are UTC?
	// Locked?
	eventData map[string]*eventBlob
}

// The rootDir must have the data/ and state/ subdirectories already, as well as the
// cluster-aliases file and probably other things, TBD.

func Open(rootDir path.Path) (*Store, error) {
}

func (s *Store) AddCluster(cluster string) {
}

// Lookup event records.
//
// `cluster` is the cluster name.  It must exist in the database, see `AddCluster`.
//
// `from` and `to` shall be UTC timestamps, they are used for their dates only, the dates are used
// to filter records coarsely.  All records whose timestamps, converted to UTC, are in that date
// range inclusively, are returned.  Records outside that date range may also be returned.
//
// `hostFilter` can be nil, otherwise this regex is matched against each host names and only data
// for host names that match are included.
//
// The returned slice has one subslice for each cached chunk, no assumptions should be made about
// the time ranges in the various chunks or even their order.  (The assumption is that subsequent
// filtering will happen and that flattening the results to one slice or sorting the data in any way
// is wasted.)
//
// The returned records shall be considered read-only.
//
// Errors are returned for bad cluster names, bad timestamps, regular expression execution failures,
// and I/O errors.

func (s *Store) EventRecords(
	cluster string,
	from, to time.Time,
	hostFilter *regexp.RegExp,
) ([][]*EventRecord, error) {
}

// Add event records.
//
// `cluster` is the cluster name.  It must exist in the database, see `AddCluster`.
//
// The ownership of the records is transfered to the store.  The records are inserted by the date
// given in their timestamps and the host names in their hostname fields.
//
// Errors are returned for bad cluster names, bad timestamps, and I/O errors.

func (s *Store) AddEventRecords(cluster string, []*EventRecord) error {
}
