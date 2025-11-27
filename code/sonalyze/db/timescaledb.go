package db

import (
	"encoding/base64"
	"context"
	"fmt"
	"maps"
	"math"
	"slices"
	"time"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
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
	var cmd, node, user string
	var cpuTime, epoch, job, numThreads, pid, ppid, residentMemory, virtualMemory pgtype.Int8
	var cpuAvg float64
	var rolledup int
	var timestamp time.Time

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "cmd, cpu_avg, cpu_time, epoch, job, node, num_threads, pid, ppid, " +
		"resident_memory, rolledup, time, user, virtual_memory"
	boxes := []any{
		&cmd, &cpuAvg, &cpuTime, &epoch, &job, &node, &numThreads, &pid, &ppid,
		&residentMemory, &rolledup, &timestamp, &user, &virtualMemory,
	}

	// Reference: ParseSamplesV0JSON()
	cluster := StringToUstr(cdb.cx.ClusterName())
	// The sonar version is currently lost in the timescaledb
	v0 := StringToUstr("0.0.0")
	unbox := func() *repr.Sample {
		return &repr.Sample{
			// TODO: Gpus, GpuPct, GpuMemPct, GpuFail - requires some kind of join
			// TODO: Cores - ditto
			Version:    v0,
			Cluster:    cluster,
			Cmd:        StringToUstr(cmd),
			CpuPct:     float32(cpuAvg),
			CpuTimeSec: uint64(cpuTime.Int),
			Epoch:      uint64(epoch.Int),
			Job:        uint32(job.Int),
			Hostname:   StringToUstr(node),
			Threads:    uint32(numThreads.Int) + 1,
			Pid:        uint32(pid.Int),
			Ppid:       uint32(ppid.Int),
			RssAnonKB:  uint64(residentMemory.Int),
			Rolledup:   uint32(rolledup),
			Timestamp:  timestamp.UTC().Unix(),
			User:       StringToUstr(user),
			CpuKB:      uint64(virtualMemory.Int),
		}
	}
	return querySlice[repr.Sample](
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, "sample_process", fields)
}

func (cdb *connectedDB) ReadNodeSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sampleBlobs [][]*repr.NodeSample, softErrors int, err error) {
	var existingEntities, runnableEntities, usedMemory pgtype.Int8
	var load1, load15, load5 float64
	var node string
	var timestamp time.Time

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "existing_entities, load1, load15, load5, node, " +
		"runnable_entities, time, used_memory"
	boxes := []any{
		&existingEntities, &load1, &load15, &load5, &node,
		&runnableEntities, &timestamp, &usedMemory,
	}

	// Reference: ParseSamplesV0JSON
	unbox := func() *repr.NodeSample {
		return &repr.NodeSample{
			ExistingEntities: uint64(existingEntities.Int),
			Hostname:         StringToUstr(node),
			Load1:            load1,
			Load5:            load5,
			Load15:           load15,
			RunnableEntities: uint64(runnableEntities.Int),
			Timestamp:        timestamp.UTC().Unix(),
			UsedMemory:       uint64(usedMemory.Int),
		}
	}
	return querySlice[repr.NodeSample](
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, "sample_system", fields)
}

func (cdb *connectedDB) ReadCpuSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.CpuSamples, softErrors int, err error) {
	var cpus []pgtype.Int8
	var node string
	var timestamp time.Time

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "cpus, node, time"
	boxes := []any{&cpus, &node, &timestamp}

	// Reference: ParseSamplesV0JSON
	unbox := func() *repr.CpuSamples {
		cpuLoad := make([]uint64, len(cpus))
		for i, n := range cpus {
			cpuLoad[i] = uint64(n.Int)
		}
		return &repr.CpuSamples{
			Hostname:  StringToUstr(node),
			Timestamp: timestamp.UTC().Unix(),
			Encoded:   repr.EncodedCpuSamplesFromValues(cpuLoad),
		}
	}
	return querySlice[repr.CpuSamples](
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, "sample_system", fields)
}

func (cdb *connectedDB) ReadGpuSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.GpuSamples, softErrors int, err error) {
	// This is a mess, because the timescaledb representation of a GPU sample does not contain the
	// node that the card was on at the time (nor the node's cluster), so we're going to have to
	// look that up somehow.  (The connectedDB has the cluster of course.)  I have filed a bug for
	// this weirdness in the data model.  It also means that we can't query the DB by cluster so the
	// standard query mechanism will not work in any case.  It may be that this API is not workable
	// for timescaledb because the assumption there is that you go cluster -> host -> card_uuid and
	// then lookup cards.  We have a host set here but it could be open... and then we'd have to
	// enumerate all cards for all hosts for the entire time period, I don't see that happening.
	// I'm guessing that the consumers could branch based on the DB type, and/or branch on the
	// distinguished error that says that the API operation is unsupported.
	return nil, 0, NewAPINotSupportedError("GPU samples")
}

func (cdb *connectedDB) ReadSysinfoNodeData(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sysinfoBlobs [][]*repr.SysinfoNodeData, softErrors int, err error) {
	var architecture, cluster, cpuModel, node, osName, osRelease, topoSvg, topoText string
	var coresPerSocket, memory, sockets, threadsPerCore pgtype.Int8
	var timestamp time.Time
	var distances []int
	var cards []string

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "architecture, cards, cluster, cores_per_socket, cpu_model, distances, memory, " +
		"node, os_name, os_release, sockets, threads_per_core, time, topo_svg, topo_text"
	boxes := []any{
		&architecture, &cards, &cluster, &coresPerSocket, &cpuModel, &distances, &memory,
		&node, &osName, &osRelease, &sockets, &threadsPerCore, &timestamp, &topoSvg, &topoText,
	}

	// Reference: ParseSysinfoV0JSON
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
		// Arguably a database bug that these are kept encoded in the database.
		if topoSvg != "" {
			dst := make([]byte, base64.StdEncoding.DecodedLen(len(topoSvg)))
			n, err := base64.StdEncoding.Decode(dst, []byte(topoSvg))
			if err != nil {
				softErrors++
				topoSvg = ""
			} else {
				topoSvg = string(dst[:n])
			}
		}
		if topoText != "" {
			dst := make([]byte, base64.StdEncoding.DecodedLen(len(topoText)))
			n, err := base64.StdEncoding.Decode(dst, []byte(topoText))
			if err != nil {
				softErrors++
				topoText = ""
			} else {
				topoText = string(dst[:n])
			}
		}
		return &repr.SysinfoNodeData{
			// Useful to keep this in the same order as the ones above
			Architecture:   architecture,
			Cards:          cards,
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
			TopoSVG:        topoSvg,
			TopoText:       topoText,
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
	// TODO: This is more complicated!  The DB stores what it perceives to be static card info in a
	// separate table, sysinfo_gpu_card.  That will need to be joined to sysinfo_gpu_card_config
	// here (by UUID) to get the full story.
	var address, driver, firmware, node, uuid string
	var index int
	var maxCeClock, maxMemoryClock, maxPowerLimit, minPowerLimit, powerLimit pgtype.Int8
	var timestamp time.Time

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "address, driver, firmware, index, max_ce_clock, max_memory_clock, " +
		"max_power_limit, min_power_limit, node, power_limit, time, uuid"
	boxes := []any{
		&address, &driver, &firmware, &index, &maxCeClock, &maxMemoryClock,
		&maxPowerLimit, &minPowerLimit, &node, &powerLimit, &timestamp, &uuid,
	}

	// Reference: ParseSysinfoV0JSON
	unbox := func() *repr.SysinfoCardData {
		return &repr.SysinfoCardData{
			Time: timestamp.Format(time.RFC3339),
			Node: node,
			SysinfoGpuCard: &newfmt.SysinfoGpuCard{
				Address:        address,
				Driver:         driver,
				Firmware:       firmware,
				Index:          uint64(index),
				MaxCEClock:     uint64(maxCeClock.Int),
				MaxMemoryClock: uint64(maxMemoryClock.Int),
				MaxPowerLimit:  uint64(maxPowerLimit.Int),
				MinPowerLimit:  uint64(minPowerLimit.Int),
				PowerLimit:     uint64(powerLimit.Int),
				UUID:           uuid,
			},
		}
	}

	return querySlice[repr.SysinfoCardData](
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, "sysinfo_gpu_card_config", fields)
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
	unbox func() *T,
	table, fields string,
) (finalRows [][]*T, softErrors int, err error) {
	// TODO: time span and hosts obviously
	// TODO: what about quoting the cluster name?
	// TODO: literal sql is an antipattern, there must be something better?  Is quoting automatic?
	qstr := "SELECT " + fields + " FROM " + table + " WHERE cluster=$1"
	qarg := []any{cdb.cx.ClusterName()}
	rows, err := cdb.theDB.connection.Query(context.Background(), qstr, qarg...)
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
