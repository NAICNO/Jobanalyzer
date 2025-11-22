package db

import (
	"context"
	"fmt"
	"maps"
	"math"
	"slices"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
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
	connection *pgx.Conn
}

func OpenDatabaseURI(databaseURI string) (*databaseConnection, error) {
	connection, err := pgx.Connect(context.Background(), databaseURI)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	//defer conn.Close(context.Background())
	return &databaseConnection{connection}, nil
}

func (cdb *databaseConnection) EnumerateClusters() ([]string, error) {
	rows, err := cdb.connection.Query(context.Background(), "SELECT cluster FROM cluster_attributes")
	if err != nil {
		return nil, err
	}
	rawClusters, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}
	// Workaround for https://github.com/2maz/slurm-monitor/issues/17 which we can remove soon:
	// cluster names can be duplicated.
	uniqueClusters := make(map[string]bool)
	for _, c := range rawClusters {
		uniqueClusters[c] = true
	}
	return slices.Collect(maps.Keys(uniqueClusters)), nil
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
	// TODO: topo_svg, topo_text
	var architecture, cluster, cpuModel, node, osName, osRelease string
	var coresPerSocket, memory, sockets, threadsPerCore pgtype.Int8
	var timestamp time.Time
	var distances []int

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields :=
		"architecture, cluster, cores_per_socket, cpu_model, distances, memory, " +
		"node, os_name, os_release, sockets, threads_per_core, time"
	boxes := []any{
		&architecture, &cluster, &coresPerSocket, &cpuModel, &distances, &memory,
		&node, &osName, &osRelease, &sockets, &threadsPerCore, &timestamp,
	}

	unbox := func() *repr.SysinfoNodeData {
		dside := int(math.Sqrt(float64(len(distances))))
		ds := make([][]uint64, dside)
		k := 0
		for i := range dside {
			ds[i] = make([]uint64, dside)
			for j := range dside {
				ds[i][j] = uint64(distances[k])
				j++
				k++
			}
		}
		return &repr.SysinfoNodeData{
			// Useful to keep this in the same order as the ones above
			Architecture:   architecture,
			Cluster:        cluster,
			CoresPerSocket: uint64(coresPerSocket.Int),
			CpuModel:       cpuModel,
			Distances:      ds,
			Memory:         uint64(memory.Int),
			Node:           node,
			OsName:         osName,
			OsRelease:      osRelease,
			Sockets:        uint64(sockets.Int),
			ThreadsPerCore: uint64(threadsPerCore.Int),
			Time:           timestamp.Format(time.RFC3339),
		}
	}

	return querySlice[repr.SysinfoNodeData](
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, "sysinfo_attributes", fields)
}

func (cdb *connectedDB) ReadSysinfoCardData(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sysinfoBlobs [][]*repr.SysinfoCardData, softErrors int, err error) {
	return
	//panic("NYI")
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

func querySlice[T any](
	cdb *connectedDB,
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
	boxes []any,
	unbox func () *T,
	table, fields string,
) (finalRows [][]*T, softErrors int, err error) {
	// TODO: time span and hosts obviously
	// TODO: what about quoting the cluster name?
	// TODO: literal sql is an antipattern, there must be something better?  Is quoting automatic?
	rows, err := cdb.theDB.connection.Query(
		context.Background(),
		"SELECT "+fields+" FROM "+table+" WHERE cluster=$1",
		cdb.cx.ClusterName(),
	)
	if err != nil {
		return
	}
	dataRows := make([]*T, 0)
	_, err = pgx.ForEachRow(rows, boxes, func() error {
		dataRows = append(dataRows, unbox())
		return nil
	})
	if err != nil {
		return
	}
	finalRows = [][]*T{dataRows}
	return
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
