package sonarlog

import (
	"time"

	"go-utils/hostglob"
	"sonalyze/db"
)

func init() {
	// Set up postprocessing of samples as they are read from file, before caching them.  This is
	// currently very basic, just enough to ensure that the db.Samples are read-only after reading.
	db.SampleRectifier = standardSampleRectifier
}

// TODO: OPTIMIZEME: We can do more here now that we implement caching.  It will be possible (indeed
// desirable) to preprocess the cached data as much as possible so that all we need to do in
// createInputStreams (or later) is stitch streams together and resolve issues at the joins, and
// finally compute the computed (context-depenent) fields, currently only CpuUtilPct.
//
// In particular, it's possible that ReadSampleStreamsAndMaybeBounds should not return an
// InputStreamSet but a set of those, and let later processing take care of stitching and filtering
// at the same time.  The data thus returned would be read-only.

func ReadSampleStreamsAndMaybeBounds(
	c db.SampleCluster,
	fromDate, toDate time.Time,
	hostGlobber *hostglob.HostGlobber,
	recordFilter db.SampleFilter,
	wantBounds bool,
	verbose bool,
) (
	streams InputStreamSet,
	bounds Timebounds,
	read, dropped int,
	err error,
) {
	sampleBlobs, dropped, err := c.ReadSamples(fromDate, toDate, hostGlobber, verbose)
	if err != nil {
		return
	}
	for _, samples := range sampleBlobs {
		read += len(samples)
	}
	streams, bounds = createInputStreams(sampleBlobs, recordFilter, wantBounds)
	return
}

func ReadLoadDataStreams(
	c db.SampleCluster,
	fromDate, toDate time.Time,
	hostGlobber *hostglob.HostGlobber,
	verbose bool,
) (
	streams LoadDataSet,
	bounds Timebounds,
	read, dropped int,
	err error,
) {
	// Read and establish invariants

	dataBlobs, dropped, err := c.ReadLoadData(fromDate, toDate, hostGlobber, verbose)
	if err != nil {
		return
	}
	for _, data := range dataBlobs {
		read += len(data)
	}
	streams, bounds, errors := rectifyLoadData(dataBlobs)
	dropped += errors
	return
}
