package db

import (
	"errors"
	"time"

	. "sonalyze/common"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

// Rough plan:
//
// - Using sonalyze for ingestion into a real database is not an important use case at this point,
//   so it's OK for the insertion methods on the DB returned from OpenConnectedDB to do nothing or
//   to panic.  It is sufficient for the daemon to run without -kafka and with -no-add for this
//   functionality not to be touched.
//
// - Hence, we implement the read functionality only: the methods for reading the various data
//   streams from the database, ie, DataProvider.
//
// - For the initial cut, we will read raw data from the database every time, no caching in
//   Sonalyze.  Only if this is a performance issue will we add caching.
//
// - Initial implementation should emphasize EnumerateClusters() and ReadSysinfoNodeData(), to get
//   something going (maybe need to return empty sets from ReadSysinfoCardData to avoid panic).

type databaseConnection struct {
	// stuff
}

func OpenDatabaseURI(databaseURI string) (*databaseConnection, error) {
	return nil, errors.New("No database connection yet")
}

func (cdb *databaseConnection) EnumerateClusters() ([]string, error) {
	// SELECT cluster FROM cluster_attributes ;
	return nil, errors.New("Database connection not open")
}

type connectedDB struct {
	theDB *databaseConnection
	cx    types.Context
}

var _ = AppendablePersistentDataProvider((*connectedDB)(nil))

func OpenConnectedDB(cx types.Context) AppendablePersistentDataProvider {
	theDB := cx.ConnectedDB().(*databaseConnection)
	return &connectedDB{theDB, cx}
}

func (cdb *connectedDB) ReadProcessSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sampleBlobs [][]*repr.Sample, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadNodeSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sampleBlobs [][]*repr.NodeSample, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadCpuSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.CpuSamples, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadGpuSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.GpuSamples, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadSysinfoNodeData(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sysinfoBlobs [][]*repr.SysinfoNodeData, softErrors int, err error) {
	// clusterName := cdb.cx.ClusterName()
	// SELECT * FROM sysinfo_attributes WHERE cluster = ${clusterName} AND time >= ${fromDate} AND time <= ${toDate} ;
	// Oh boy, filtering by hosts will be interesting
	panic("NYI")
}

func (cdb *connectedDB) ReadSysinfoCardData(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sysinfoBlobs [][]*repr.SysinfoCardData, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadSacctData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.SacctInfo, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadCluzterAttributeData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterAttributes, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadCluzterPartitionData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterPartitions, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) ReadCluzterNodeData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterNodes, softErrors int, err error) {
	panic("NYI")
}

func (cdb *connectedDB) AppendSamplesAsync(ty DataReprType, host, timestamp string, payload any) error {
	panic("NYI")
}

func (cdb *connectedDB) AppendSysinfoAsync(ty DataReprType, host, timestamp string, payload any) error {
	panic("NYI")
}

func (cdb *connectedDB) AppendSlurmSacctAsync(ty DataReprType, timestamp string, payload any) error {
	panic("NYI")
}

func (cdb *connectedDB) AppendCluzterAsync(ty DataReprType, timestamp string, payload any) error {
	panic("NYI")
}

func (cdb *connectedDB) FlushAsync() {
	// Do nothing
}

func (cdb *connectedDB) Close() error {
	// Do nothing
	return nil
}
