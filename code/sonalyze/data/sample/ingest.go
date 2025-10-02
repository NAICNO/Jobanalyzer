package sample

import (
	"time"

	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/db/filedb"
)

func init() {
	// Set up postprocessing of samples as they are read from file, before caching them.  This is
	// currently very basic, just enough to ensure that the db.Samples are read-only after reading.
	// Note the rectifier is not applied to data coming from non-file sources.
	filedb.SampleRectifier = standardSampleRectifier
}

// Read data and bucket them in InputStreams.  The caller receives ownership of the InputStreamSet.
// The spines of the data structures may be modified, but the ultimate repr.Sample value is owned by
// the database and is read-only.  Any augmentation must be added in the wrapping sample.Sample
// object.
//
// TODO: OPTIMIZEME: We can do more here now that we implement caching.  It will be possible (indeed
// desirable) to preprocess the cached data as much as possible so that all we need to do in
// createInputStreams (or later) is stitch streams together and resolve issues at the joins, and
// finally compute the computed (context-depenent) fields, currently only CpuUtilPct.
//
// In particular, it's possible that ReadSampleStreamsAndMaybeBounds should not return an
// InputStreamSet but a set of those, and let later processing take care of stitching and filtering
// at the same time.  The data thus returned would be read-only.

func ReadSampleStreamsAndMaybeBounds(
	c db.ProcessSampleDataProvider,
	fromDate, toDate time.Time,
	hostGlobber *Hosts,
	recordFilter *SampleFilter,
	wantBounds bool,
	verbose bool,
) (
	streams InputStreamSet,
	bounds Timebounds,
	read, dropped int,
	err error,
) {
	sampleBlobs, dropped, err := c.ReadProcessSamples(fromDate, toDate, hostGlobber, verbose)
	if err != nil {
		return
	}
	for _, samples := range sampleBlobs {
		read += len(samples)
	}
	streams, bounds = createInputStreams(sampleBlobs, recordFilter, wantBounds)
	computePerSampleFields(streams)
	return
}
