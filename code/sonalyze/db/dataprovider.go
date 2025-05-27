// Data providers - abstract interface to a data store.  There are several types of store, see
// go.doc in this directory.

package db

import (
	"time"

	"go-utils/config"

	. "sonalyze/common"
	"sonalyze/db/filedb"
	"sonalyze/db/repr"
)

// DataReprType enumerates the representations of data coming in from external sources.
type DataReprType = filedb.FileAttr

const (
	DataSysinfoOldJSON = filedb.FileSysinfoOldJSON
	DataSampleCSV      = filedb.FileSampleCSV
	DataSlurmCSV       = filedb.FileSlurmCSV
	DataSampleV0JSON   = filedb.FileSampleV0JSON
	DataSysinfoV0JSON  = filedb.FileSysinfoV0JSON
	DataSlurmV0JSON    = filedb.FileSlurmV0JSON
	DataCluzterV0JSON  = filedb.FileCluzterV0JSON
)

// The readers return slices of data that may be shared with the database.  The inner slices of the
// result, and the records they point to, must not be mutated in ANY way.
//
// Time bounds given to readers and file name extractors must be UTC.

// SampleDataProvider is a reader for data in "Sample" data streams, which carry both process
// samples, CPU load data, and GPU utilization data.
type SampleDataProvider interface {
	ReadSamples(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (sampleBlobs [][]*repr.Sample, softErrors int, err error)

	ReadLoadData(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (dataBlobs [][]*repr.LoadDatum, softErrors int, err error)

	ReadGpuData(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (dataBlobs [][]*repr.GpuDatum, softErrors int, err error)
}

// SysinfoDataProvider is a reader for data in "Sysinfo" data streams.
type SysinfoDataProvider interface {
	ReadSysinfoData(
		fromDate, toDate time.Time,
		hosts *Hosts,
		verbose bool,
	) (sysinfoBlobs [][]*repr.SysinfoData, softErrors int, err error)
}

// SacctDataProvider is a reader for Slurm/sacct data in "Jobs" data streams.
type SacctDataProvider interface {
	ReadSacctData(
		fromDate, toDate time.Time,
		verbose bool,
	) (recordBlobs [][]*repr.SacctInfo, softErrors int, err error)
}

// CluzterDataProvider is a reader for Slurm/sinfo data in "Cluster" data streams.  The name
// "Cluzter" is used at present to disambiguate with several other uses of "Cluster".
type CluzterDataProvider interface {
	ReadCluzterData(
		fromDate, toDate time.Time,
		verbose bool,
	) (recordBlobs [][]*repr.CluzterInfo, softErrors int, err error)
}

// PersistentDataProvider provides all data types, and carries a cluster configuration with it.
type PersistentDataProvider interface {
	Config() *config.ClusterConfig

	SampleDataProvider
	SysinfoDataProvider
	SacctDataProvider
	CluzterDataProvider
}

// AppendabePersistentDataProvider allows the data stores to be extended in a safe way.
//
// `payload` is string or []byte, exclusively.  Each should in general be a single record.  The
// payload may optionally be terminated with \n to indicate end-of-record; any embedded \n are
// technically considered part of the record and is only allowed if the record format allows
// that (JSON does, CSV does not).
type AppendablePersistentDataProvider interface {
	PersistentDataProvider

	AppendSamplesAsync(ty DataReprType, host, timestamp string, payload any) error
	AppendSysinfoAsync(ty DataReprType, host, timestamp string, payload any) error
	AppendSlurmSacctAsync(ty DataReprType, timestamp string, payload any) error
	AppendCluzterAsync(ty DataReprType, timestamp string, payload any) error

	FlushAsync()
	Close() error
}

// Databases opened on sets of files or on directory trees can also provide the names of the files
// they operate on.

// SampleFilenameProvider returns the list of "Sample" data files in used by the database.
type SampleFilenameProvider interface {
	SampleFilenames(
		fromDate, toDate time.Time,
		hosts *Hosts,
	) ([]string, error)
}

// SysinfoFilenameProvider returns the list of "Sysinfo" data files in used by the database.
type SysinfoFilenameProvider interface {
	SysinfoFilenames(
		fromDate, toDate time.Time,
		hosts *Hosts,
	) ([]string, error)
}

// SacctFilenameProvider returns the list of "Jobs" data files in used by the database.
type SacctFilenameProvider interface {
	SacctFilenames(
		fromDate, toDate time.Time,
	) ([]string, error)
}

// CluzterFilenameProvider returns the list of "Cluster" data files in used by the database.
type CluzterFilenameProvider interface {
	CluzterFilenames(
		fromDate, toDate time.Time,
	) ([]string, error)
}
