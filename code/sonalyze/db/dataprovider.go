// Data providers - abstract interface to a data store.  There are several types of store, see
// doc.go in this directory.

package db

import (
	"time"

	"sonalyze/db/filedb"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

// DataReprType enumerates the representations of data coming in from external sources and is used
// to communicate the representation to the data-insertion methods.
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

// The readers in the SampleDataProvider return slices of data that may be shared with the database.
// The inner slices of the result, and the records they point to, must not be mutated in ANY way.
// The outer slices are however freshly created and may be updated by the caller.
//
// Time bounds given to readers and file name extractors must be UTC.

// SampleDataProvider is a reader for data in "Sample" data streams, which carry process samples
// and node, disk, CPU and GPU utilization data.
type SampleDataProvider interface {
	ProcessSampleDataProvider
	NodeSampleDataProvider
	DiskSampleDataProvider
	CpuSampleDataProvider
	GpuSampleDataProvider
}

type APINotSupportedError struct {
	message string
}

func NewAPINotSupportedError(message string) error {
	return &APINotSupportedError{message}
}

func (e *APINotSupportedError) Error() string {
	return "API not supported: " + e.message
}

type ProcessSampleDataProvider interface {
	ReadProcessSamples(
		filter types.DataProviderFilter,
	) (sampleBlobs [][]*repr.Sample, softErrors int, err error)
}

type NodeSampleDataProvider interface {
	ReadNodeSamples(
		filter types.DataProviderFilter,
	) (sampleBlobs [][]*repr.NodeSample, softErrors int, err error)
}

type DiskSampleDataProvider interface {
	ReadDiskSamples(
		filter types.DataProviderFilter,
	) (dataBlobs [][]*repr.DiskSample, softErrors int, err error)
}

type CpuSampleDataProvider interface {
	ReadCpuSamples(
		filter types.DataProviderFilter,
	) (dataBlobs [][]*repr.CpuSamples, softErrors int, err error)
}

type GpuSampleDataProvider interface {
	ReadGpuSamples(
		filter types.DataProviderFilter,
	) (dataBlobs [][]*repr.GpuSamples, softErrors int, err error)
}

// SysinfoDataProvider is a reader for data in "Sysinfo" data streams.
type SysinfoDataProvider interface {
	ReadSysinfoNodeData(
		filter types.DataProviderFilter,
	) (sysinfoBlobs [][]*repr.SysinfoNodeData, softErrors int, err error)

	ReadSysinfoCardData(
		filter types.DataProviderFilter,
	) (sysinfoBlobs [][]*repr.SysinfoCardData, softErrors int, err error)
}

// SacctDataProvider is a reader for Slurm/sacct data in "Jobs" data streams.
type SacctDataProvider interface {
	ReadSacctData(
		filter types.DataProviderFilter,
	) (recordBlobs [][]*repr.SacctInfo, softErrors int, err error)
}

// CluzterDataProvider is a reader for Slurm/sinfo data in "Cluster" data streams.  The name
// "Cluzter" is used at present to disambiguate with several other uses of "Cluster" in the database
// code.
type CluzterDataProvider interface {
	ReadCluzterAttributeData(
		filter types.DataProviderFilter,
	) (recordBlobs [][]*repr.CluzterAttributes, softErrors int, err error)

	ReadCluzterPartitionData(
		filter types.DataProviderFilter,
	) (recordBlobs [][]*repr.CluzterPartitions, softErrors int, err error)

	ReadCluzterNodeData(
		filter types.DataProviderFilter,
	) (recordBlobs [][]*repr.CluzterNodes, softErrors int, err error)
}

// MetaDataProvider computes (and sometimes caches) data about the data in the database.
type MetaDataProvider interface {
	MinTime(soft bool) (time.Time, error)
	MaxTime(soft bool) (time.Time, error)
}

// DataProvider provides all data types.
type DataProvider interface {
	SampleDataProvider
	SysinfoDataProvider
	SacctDataProvider
	CluzterDataProvider
	MetaDataProvider
}

// PersistentDataProvider is backed by persistent storage in some way, and provides all data types.
type PersistentDataProvider = DataProvider

// AppendabePersistentDataProvider allows persistent data stores to be extended in a safe way.
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
	SampleFilenames(filter types.DataProviderFilter) ([]string, error)
}

// SysinfoFilenameProvider returns the list of "Sysinfo" data files in used by the database.
type SysinfoFilenameProvider interface {
	SysinfoFilenames(filter types.DataProviderFilter) ([]string, error)
}

// SacctFilenameProvider returns the list of "Jobs" data files in used by the database.
type SacctFilenameProvider interface {
	SacctFilenames(filter types.DataProviderFilter) ([]string, error)
}

// CluzterFilenameProvider returns the list of "Cluster" data files in used by the database.
type CluzterFilenameProvider interface {
	CluzterFilenames(filter types.DataProviderFilter) ([]string, error)
}
