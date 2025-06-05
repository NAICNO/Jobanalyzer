package sample

import (
	. "sonalyze/common"
	"sonalyze/db/errs"
	"sonalyze/db/repr"
)

// Code would be simpler if we would embed the db.Sample here but that structure is currently (and
// probably forever) large enough that memory spikes would be a real concern.  Hence the
// indirection.

type Sample struct {
	*repr.Sample         // Read-only (adjusted) log data
	CpuUtilPct   float32 // Computed from a concrete selection
}

// A sample stream is just a list of samples.

type SampleStream []Sample

// A bag of streams.  The constraints on the individual streams in terms of uniqueness and so on
// depends on how they were merged and are not implied by the type.

type SampleStreams []*SampleStream

var (
	// MT: Constant after initialization; immutable
	BadTimestampErr = errs.BadTimestampErr
)

// The InputStreamKey is (hostname, stream-id, cmd), where the stream-id is defined below; it is
// meaningful only for non-merged streams.
//
// An InputStreamSet maps a InputStreamKey to a SampleStream pertinent to that key.  It is named as
// it is because the InputStreamKey is meaningful only for non-merged streams.

type InputStreamKey struct {
	Host     Ustr
	StreamId uint32
	Cmd      Ustr
}

// The streams are heap-allocated so that we can update them without also updating the map.
//
// After postprocessing, there are some important invariants on the records that make up an input
// stream in addition to them having the same key:
//
// - the vector is sorted ascending by timestamp
// - no two adjacent timestamps are the same
// - the Decoded arrays within a given Data array all have the same length

type InputStreamSet map[InputStreamKey]*SampleStream
