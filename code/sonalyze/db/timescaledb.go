// This is a reader-only interface to the timescaledb, allowing it to be used as the primitive data
// store for sonalyze.  This in turn allows all of sonalyze's application logic (stream merging etc)
// to be applied to data stored in timescaledb so that we don't have to duplicate it in the Python
// code.
//
// The interface is read-only because ingestion into timescaledb is handled by external ingestion
// code, as part of slurm-monitor.  The insertion methods on the DB returned from OpenConnectedDB
// will panic, do not call them.  It is sufficient for the daemon to run without -kafka and with
// -no-add for this functionality not to be touched.
//
// Here we read raw data from the database every time, no caching in Sonalyze.  Only if this is a
// performance issue will we add caching.  I do not expect this to happen.  Instead, I expect there
// to be higher-level accessors to the data added to data/ that will make caching unnecessary.
//
// Nullable fields!
//
// As of 2026-02-20 the following fields can be null (the 2nd column is their status in this code):
//
//  sample_gpu.index                             ok
//  sample_process.in_container                  ok
//  sample_slurm_job.array_job_id                ok
//  sample_slurm_job.array_task_id               ok
//  sample_slurm_job.start_time                  ok
//  sample_slurm_job.end_time                    ok
//  sample_slurm_job.exit_code                   ok
//  sample_slurm_job.nodes                       probably ok (slice)
//  sample_slurm_job.gres_detail                 not used
//  sample_slurm_job.requested_resources         ok
//  sample_slurm_job.allocated_resources         ok
//  sample_slurm_job.minimum_cpus_per_node       ok
//  sample_system.boot                           ok
//  sysinfo_attributes.topo_svg                  ok
//  sysinfo_attributes.topo_text                 ok
//  sysinfo_attributes.numa_nodes                ok
//  sysinfo_attributes.distances                 probably ok (slice)
//
// It appears that the PSQL layer allows non-nullable containers (eg *string) to be passed to
// receive string values from nullable string fields, but will error out if the field has a null
// value.  So we'll need to pass **string in this case.  Ditto for the other types, unless there are
// PSQL layer types that have a Present flag and can handle this directly.
//
// The cliche is to allocate a variable of the pointer type:
//
//    var requestedResourcesp *string
//
// and then pass &requestedResourcesp to the query, and when we come back, the value is stored in
// *requestedResourcesp if not nil.  If so, the box will have been dynamically allocated, no matter
// what the initial value of requestedResourcesp.  Now we generally want to do:
//
//    var requestedResources string
//    if requestedResourcesp != nil {
//        requestedResources = *requestedResourcesp
//    }
//
// and then our null-or-not value becomes an empty-string-or-not value.
//
// Computed fields can sometimes become NULL, notably sum() over an empty set of rows, and must be
// handled the same way.

package db

import (
	"context"
	"encoding/base64"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go-utils/gpuset"
	"go-utils/hostglob"
	. "sonalyze/common"
	"sonalyze/db/repr"
	"sonalyze/db/types"
	"sonalyze/db/util"
)

// Allow no more than this many query parameters from each open-ended set (users, jobs, node, ...)
// to avoid DoS.
const dosCutoff = 100

// The current structure of sonalyze ensures that there is one databaseConnection globally, and it
// is really never closed.  (There is one connection, it is attached to every cluster when the data
// store is opened, and the cluster table is never cleared out during normal operations.)  In
// principle, there should be a finalizer on databaseConnection that closes the underlying pgx.Conn
// but it would never be called the way things are.
type databaseConnection struct {
	// The connection is not thread-safe.  Use the Query method to perform a query safely, it will
	// acquire a mutex around the connection use (or it could manage a connection pool for better
	// multi-threaded access).
	connection *pgx.Conn
	lock       sync.Mutex
}

func (cdb *databaseConnection) Query(cx context.Context, q string, arg ...any) (pgx.Rows, error) {
	cdb.lock.Lock()
	defer cdb.lock.Unlock()
	return cdb.connection.Query(cx, q, arg...)
}

func (cdb *databaseConnection) QueryRowAndScan(cx context.Context, q string, args []any, slots []any) error {
	cdb.lock.Lock()
	defer cdb.lock.Unlock()
	return cdb.connection.QueryRow(cx, q, args...).Scan(slots...)
}

func OpenDatabaseURI(databaseURI string) (*databaseConnection, error) {
	connection, err := pgx.Connect(context.Background(), databaseURI)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	return &databaseConnection{connection: connection}, nil
}

func (cdb *databaseConnection) EnumerateClusters() ([]string, error) {
	rows, err := cdb.Query(
		context.Background(),
		"SELECT cluster FROM cluster_attributes GROUP BY cluster",
	)
	if err != nil {
		return nil, err
	}
	rawClusters, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}
	// Workaround for https://github.com/2maz/slurm-monitor/issues/17 which we can remove soon:
	// cluster names can be duplicated.  Also, the GROUP BY should take care of it.
	uniqueClusters := make(map[string]bool)
	for _, c := range rawClusters {
		uniqueClusters[c] = true
	}
	return slices.Collect(maps.Keys(uniqueClusters)), nil
}

type connectedDB struct {
	theDB     *databaseConnection
	cx        types.Context
	timeCache *util.TimeCache
}

var _ = AppendablePersistentDataProvider((*connectedDB)(nil))

func OpenConnectedDB(cx types.Context) AppendablePersistentDataProvider {
	theDB := cx.ConnectedDB().(*databaseConnection)
	timeCache := util.NewTimeCache(makeRefillTimeCache(theDB, cx.ClusterName()))
	return &connectedDB{theDB, cx, timeCache}
}

// NOTE that the names in fields are a little bit brittle in the face of schema evolution and joins.
// They can always be written as "t1.field" and "t2.field", and it's useful to do so, because if
// they are just "field" then even if that is unambiguous at the time it's written, if a field of
// that name is added to the other table then it will become ambiguous.
//
// The field map is used for the inner/primary query only and allows table-local names to be used
// where they are not canonical (eg "user" is "user_name" in slurm data).  A missing mapping will
// cause a panic, it's the client's responsibility to get this right.

type query struct {
	types.DataProviderFilter
	table    string            // base table name for first-level selection, the result is "t1"
	join     string            // a join clause + an additional table t2 + join conditions
	fields   string            // comma-separated list of names
	boxes    []any             // in the same order as the fields
	fieldMap map[string]string // map from a "canonical" field name to local variant
}

func (cdb *connectedDB) ReadProcessSamples(
	filter types.DataProviderFilter,
) (sampleBlobs [][]*repr.Sample, softErrors int, err error) {
	var (
		cmd, node, user                                                           string
		cpuTime, epoch, job, numThreads, pid, ppid, residentMemory, virtualMemory pgtype.Int8
		dataCancelled, dataRead, dataWritten                                      pgtype.Int8
		cpuAvg, cpuUtil                                                           float64
		rolledup, gpuCount                                                        int
		timestamp                                                                 time.Time

		// Nullable, ignore NULL and treat as zero/false.
		gpuUtilp, gpuMemoryUtilp *float64
		gpuMemoryp               pgtype.Int8
		inContainerp             *bool
	)

	// Alpha order and KEEP THE FIELD AND BOX LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	t1Fields := "t1.cmd, t1.cpu_avg, t1.cpu_time, t1.cpu_util, t1.data_cancelled, t1.data_read, " +
		"t1.data_written, t1.epoch, t1.job, t1.node, t1.num_threads, t1.pid, t1.ppid, " +
		"t1.resident_memory, t1.rolledup, t1.time, t1.user, t1.virtual_memory"
	t1Boxes := []any{
		&cmd, &cpuAvg, &cpuTime, &cpuUtil, &dataCancelled, &dataRead,
		&dataWritten, &epoch, &job, &node, &numThreads, &pid, &ppid,
		&residentMemory, &rolledup, &timestamp, &user, &virtualMemory,
	}
	t2Fields := "sum(t2.gpu_memory), sum(t2.gpu_util), sum(t2.gpu_memory_util), count(t2.uuid)"
	t2Boxes := []any{&gpuMemoryp, &gpuUtilp, &gpuMemoryUtilp, &gpuCount}
	joinBy := "left join sample_process_gpu as t2 on t1.cluster = t2.cluster and t1.node = t2.node " +
		"and t1.time = t2.time and t1.pid = t2.pid and t1.job = t2.job and t1.epoch = t2.epoch " +
		"group by " + t1Fields
	q := query{
		DataProviderFilter: filter,
		table:              "sample_process",
		join:               joinBy,
		fields:             t1Fields + ", " + t2Fields,
		boxes:              append(t1Boxes, t2Boxes...),
	}

	// Reference: ParseSamplesV0JSON()
	cluster := StringToUstr(cdb.cx.ClusterName())
	// The sonar version is currently lost in the timescaledb
	v0 := StringToUstr("0.0.0")
	unbox := func() *repr.Sample {
		// gpu fields can be null / 0
		// Can't do: GpuFail, info appears lost
		// Won't do: NumCores, nobody cares, it's obsolete
		gpus := gpuset.EmptyGpuSet()
		if gpuCount > 0 {
			// Note, information about precise indices is lost
			for i := 0; i < gpuCount; i++ {
				gpus, _ = gpuset.Adjoin(gpus, 1<<i)
			}
		}
		var gpuMemory uint64
		if gpuMemoryp.Valid {
			gpuMemory = uint64(gpuMemoryp.Int64)
		}
		var gpuUtil, gpuMemoryUtil float32
		if gpuUtilp != nil {
			gpuUtil = float32(*gpuUtilp)
		}
		if gpuMemoryUtilp != nil {
			gpuMemoryUtil = float32(*gpuMemoryUtilp)
		}
		var inContainer bool
		if inContainerp != nil {
			inContainer = *inContainerp
		}
		return &repr.Sample{
			Version:           v0,
			Cluster:           cluster,
			Cmd:               StringToUstr(cmd),
			CpuPct:            float32(cpuAvg),
			CpuTimeSec:        uint64(cpuTime.Int64),
			Epoch:             uint64(epoch.Int64),
			Job:               uint32(job.Int64),
			Hostname:          StringToUstr(node),
			NumThreads:        uint32(numThreads.Int64) + 1,
			Pid:               uint64(pid.Int64),
			Ppid:              uint32(ppid.Int64),
			RssAnonKB:         uint64(residentMemory.Int64),
			Rolledup:          uint32(rolledup),
			Timestamp:         timestamp.UTC().Unix(),
			User:              StringToUstr(user),
			CpuKB:             uint64(virtualMemory.Int64),
			Gpus:              gpus,
			GpuPct:            gpuUtil,
			GpuMemPct:         gpuMemoryUtil,
			GpuKB:             gpuMemory,
			InContainer:       inContainer,
			CpuSampledUtilPct: float32(cpuUtil),
			DataReadKB:        uint64(dataRead.Int64),
			DataWrittenKB:     uint64(dataWritten.Int64),
			DataCancelledKB:   uint64(dataCancelled.Int64),
		}
	}
	return querySlice[repr.Sample](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadNodeSamples(
	filter types.DataProviderFilter,
) (sampleBlobs [][]*repr.NodeSample, softErrors int, err error) {
	var (
		existingEntities, runnableEntities, usedMemory pgtype.Int8
		load1, load15, load5                           float64
		node                                           string
		timestamp                                      time.Time

		// Nullable, ignore NULL and treat as zero
		bootp pgtype.Timestamptz
	)

	q := query{
		DataProviderFilter: filter,
		table:              "sample_system",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "boot, existing_entities, load1, load15, load5, node, " +
			"runnable_entities, time, used_memory",
		boxes: []any{
			&bootp, &existingEntities, &load1, &load15, &load5, &node,
			&runnableEntities, &timestamp, &usedMemory,
		},
	}

	// Reference: ParseSamplesV0JSON
	unbox := func() *repr.NodeSample {
		var boot time.Time
		if bootp.Valid {
			boot = bootp.Time
		}
		return &repr.NodeSample{
			ExistingEntities: uint64(existingEntities.Int64),
			Hostname:         StringToUstr(node),
			Load1:            load1,
			Load5:            load5,
			Load15:           load15,
			RunnableEntities: uint64(runnableEntities.Int64),
			Timestamp:        timestamp.UTC().Unix(),
			Boot:             boot.UTC().Unix(),
			UsedMemory:       uint64(usedMemory.Int64),
		}
	}
	return querySlice[repr.NodeSample](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadDiskSamples(
	filter types.DataProviderFilter,
) (dataBlobs [][]*repr.DiskSample, softErrors int, err error) {
	var (
		name, node                                                                    string
		timestamp                                                                     time.Time
		discards_completed, discards_merged, flush_requests_completed                 pgtype.Int8
		ios_currently_in_progress, major, minor, ms_spent_discarding                  pgtype.Int8
		ms_spent_doing_ios, ms_spent_flushing, ms_spent_reading, ms_spent_writing     pgtype.Int8
		reads_completed, reads_merged, sectors_discarded, sectors_read                pgtype.Int8
		sectors_written, weighted_ms_spent_doing_ios, writes_completed, writes_merged pgtype.Int8
	)

	q := query{
		DataProviderFilter: filter,
		table:              "sample_disk",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "discards_completed, discards_merged, flush_requests_completed, " +
			"ios_currently_in_progress, major, minor, ms_spent_discarding, " +
			"ms_spent_doing_ios, ms_spent_flushing, ms_spent_reading, ms_spent_writing, " +
			"name, node, reads_completed, reads_merged, sectors_discarded, sectors_read, " +
			"sectors_written, time, weighted_ms_spent_doing_ios, writes_completed, writes_merged",
		boxes: []any{&discards_completed, &discards_merged, &flush_requests_completed,
			&ios_currently_in_progress, &major, &minor, &ms_spent_discarding,
			&ms_spent_doing_ios, &ms_spent_flushing, &ms_spent_reading, &ms_spent_writing,
			&name, &node, &reads_completed, &reads_merged, &sectors_discarded, &sectors_read,
			&sectors_written, &timestamp, &weighted_ms_spent_doing_ios, &writes_completed, &writes_merged,
		},
	}

	// Reference: ParseSamplesV0JSON
	unbox := func() *repr.DiskSample {
		return &repr.DiskSample{
			Timestamp:         timestamp.UTC().Unix(),
			Hostname:          StringToUstr(node),
			Name:              StringToUstr(name),
			Major:             uint64(major.Int64),
			Minor:             uint64(minor.Int64),
			ReadsCompleted:    uint64(reads_completed.Int64),
			ReadsMerged:       uint64(reads_merged.Int64),
			SectorsRead:       uint64(sectors_read.Int64),
			MsReading:         uint64(ms_spent_reading.Int64),
			WritesCompleted:   uint64(writes_completed.Int64),
			WritesMerged:      uint64(writes_merged.Int64),
			SectorsWritten:    uint64(sectors_written.Int64),
			MsWriting:         uint64(ms_spent_writing.Int64),
			IOsInProgress:     uint64(ios_currently_in_progress.Int64),
			MsDoingIO:         uint64(ms_spent_doing_ios.Int64),
			WeightedMsDoingIO: uint64(weighted_ms_spent_doing_ios.Int64),
			DiscardsCompleted: uint64(discards_completed.Int64),
			DiscardsMerged:    uint64(discards_merged.Int64),
			SectorsDiscarded:  uint64(sectors_discarded.Int64),
			MsDiscarding:      uint64(ms_spent_discarding.Int64),
			FlushesCompleted:  uint64(flush_requests_completed.Int64),
			MsFlushing:        uint64(ms_spent_flushing.Int64),
		}
	}
	return querySlice[repr.DiskSample](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadCpuSamples(
	filter types.DataProviderFilter,
) (dataBlobs [][]*repr.CpuSamples, softErrors int, err error) {
	var (
		cpus      []pgtype.Int8
		node      string
		timestamp time.Time
	)

	q := query{
		DataProviderFilter: filter,
		table:              "sample_system",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "cpus, node, time",
		boxes:  []any{&cpus, &node, &timestamp},
	}

	// Reference: ParseSamplesV0JSON
	unbox := func() *repr.CpuSamples {
		cpuLoad := make([]uint64, len(cpus))
		for i, n := range cpus {
			cpuLoad[i] = uint64(n.Int64)
		}
		return &repr.CpuSamples{
			Hostname:  StringToUstr(node),
			Timestamp: timestamp.UTC().Unix(),
			Encoded:   repr.EncodedCpuSamplesFromValues(cpuLoad),
		}
	}
	return querySlice[repr.CpuSamples](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadGpuSamples(
	filter types.DataProviderFilter,
) (dataBlobs [][]*repr.GpuSamples, softErrors int, err error) {
	var (
		ce_clock, ce_util, failing, fan, memory, memory_clock, memory_util pgtype.Int8
		performance_state, power, power_limit                              pgtype.Int8
		temperature                                                        int
		compute_mode, node, uuid                                           string
		timestamp                                                          time.Time

		// Nullable, ignore NULL and treat as zero
		indexp *int
	)

	// Here we must start with sysinfo_gpu_card_config as to be able to filter cards by cluster and
	// node, but once that's done we're mostly interested in data from sample_gpu.  (It's a shame
	// that we have to do all of this just because sample_gpu does not carry cluster and node.)
	//
	// In age(t1,t2) we compute t1-t2 as an interval, ie, this is positive when t1 is newer.  For
	// this query, we want the latest available sample at the time of for each sysinfo.  Since
	// samples are always relatively frequent the interval can usually be quite small, but it is
	// a concern that the interval is hardcoded.
	//
	// We need the extra constraint on t2.time or we'll get records through the current time.
	extra := ""
	if !filter.ToDate.IsZero() {
		extra = "t2.time < " + toDateName(filter.FromDate, filter.ToDate) + " and "
	}
	q := query{
		DataProviderFilter: filter,
		table:              "sysinfo_gpu_card_config",
		join: "join sample_gpu as t2 on t1.uuid = t2.uuid and " + extra +
			"age(t1.time, t2.time) < interval '15 minutes'",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "t2.ce_clock, t2.ce_util, t2.compute_mode, t2.failing, t2.fan, t2.index, " +
			"t2.memory, t2.memory_clock, t2.memory_util, t1.node, t2.performance_state, " +
			"t2.power, t2.power_limit, t2.temperature, t2.time, t2.uuid",
		boxes: []any{
			&ce_clock, &ce_util, &compute_mode, &failing, &fan, &indexp,
			&memory, &memory_clock, &memory_util, &node, &performance_state,
			&power, &power_limit, &temperature, &timestamp, &uuid,
		},
	}

	// Reference: ParseSamplesV0JSON
	unbox := func() *repr.GpuSamples {
		var index int
		if indexp != nil {
			index = *indexp
		}
		data := repr.PerGpuSample{
			Attr: repr.GpuHasUuid | repr.GpuHasComputeMode | repr.GpuHasUtil | repr.GpuHasFailing,
			SampleGpu: &newfmt.SampleGpu{
				Index:            uint64(index),
				UUID:             newfmt.NonemptyString(uuid),
				Failing:          uint64(failing.Int64),
				Fan:              uint64(fan.Int64),
				ComputeMode:      compute_mode,
				PerformanceState: newfmt.ExtendedUint(performance_state.Int64),
				Memory:           uint64(memory.Int64),
				CEUtil:           uint64(ce_util.Int64),
				MemoryUtil:       uint64(memory_util.Int64),
				Temperature:      int64(temperature),
				Power:            uint64(power.Int64),
				PowerLimit:       uint64(power_limit.Int64),
				CEClock:          uint64(ce_clock.Int64),
				MemoryClock:      uint64(memory_clock.Int64),
			},
		}
		return &repr.GpuSamples{
			Hostname:  StringToUstr(node),
			Timestamp: timestamp.UTC().Unix(),
			Encoded:   repr.EncodedGpuSamplesFromValues([]repr.PerGpuSample{data}),
		}
	}
	return querySlice[repr.GpuSamples](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadSysinfoNodeData(
	filter types.DataProviderFilter,
) (sysinfoBlobs [][]*repr.SysinfoNodeData, softErrors int, err error) {
	var (
		architecture, cluster, cpuModel, node, osName, osRelease      string
		coresPerSocket, memory, numaNodesBox, sockets, threadsPerCore pgtype.Int8
		timestamp                                                     time.Time
		distances                                                     []int
		cards                                                         []string

		// Nullable, ignore NULL and treat as empty string
		topoSvgp, topoTextp *string
	)

	q := query{
		DataProviderFilter: filter,
		table:              "sysinfo_attributes",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "architecture, cards, cluster, cores_per_socket, cpu_model, distances, memory, " +
			"node, numa_nodes, os_name, os_release, sockets, threads_per_core, time, topo_svg, topo_text",
		boxes: []any{
			&architecture, &cards, &cluster, &coresPerSocket, &cpuModel, &distances, &memory,
			&node, &numaNodesBox, &osName, &osRelease, &sockets, &threadsPerCore, &timestamp, &topoSvgp, &topoTextp,
		},
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
		var numaNodes uint64
		if numaNodesBox.Valid {
			numaNodes = uint64(numaNodesBox.Int64)
		}
		var topoSvg string
		if topoSvgp != nil {
			topoSvg = *topoSvgp
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
		var topoText string
		if topoTextp != nil {
			topoText = *topoTextp
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
			CoresPerSocket: uint64(coresPerSocket.Int64),
			CpuModel:       cpuModel,
			Distances:      ds,
			Memory:         uint64(memory.Int64),
			Node:           node,
			OsName:         osName,
			OsRelease:      osRelease,
			NumaNodes:      numaNodes,
			Sockets:        uint64(sockets.Int64),
			ThreadsPerCore: uint64(threadsPerCore.Int64),
			Time:           timestamp.Format(time.RFC3339),
			TopoSVG:        topoSvg,
			TopoText:       topoText,
		}
	}

	return querySlice[repr.SysinfoNodeData](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadSysinfoCardData(
	filter types.DataProviderFilter,
) (sysinfoBlobs [][]*repr.SysinfoCardData, softErrors int, err error) {
	var (
		address, architecture, driver, firmware, manufacturer, model, node, uuid     string
		index                                                                        int
		maxCeClock, maxMemoryClock, maxPowerLimit, memory, minPowerLimit, powerLimit pgtype.Int8
		timestamp                                                                    time.Time
	)

	q := query{
		DataProviderFilter: filter,
		// The DB stores what it perceives to be static card info in a separate table,
		// sysinfo_gpu_card.  That needs to be joined to sysinfo_gpu_card_config here (by UUID) to
		// get the full story.
		table: "sysinfo_gpu_card_config",
		join:  "join sysinfo_gpu_card t2 on t1.uuid = t2.uuid",
		// Alpha field name order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "address, architecture, driver, firmware, index, manufacturer, max_ce_clock, max_memory_clock, " +
			"max_power_limit, memory, min_power_limit, model, node, power_limit, time, t1.uuid",
		boxes: []any{
			&address, &architecture, &driver, &firmware, &index, &manufacturer, &maxCeClock, &maxMemoryClock,
			&maxPowerLimit, &memory, &minPowerLimit, &model, &node, &powerLimit, &timestamp, &uuid,
		},
	}

	// Reference: ParseSysinfoV0JSON
	unbox := func() *repr.SysinfoCardData {
		return &repr.SysinfoCardData{
			Time: timestamp.Format(time.RFC3339),
			Node: node,
			SysinfoGpuCard: &newfmt.SysinfoGpuCard{
				Address:        address,
				Architecture:   architecture,
				Driver:         driver,
				Firmware:       firmware,
				Index:          uint64(index),
				Manufacturer:   manufacturer,
				MaxCEClock:     uint64(maxCeClock.Int64),
				MaxMemoryClock: uint64(maxMemoryClock.Int64),
				MaxPowerLimit:  uint64(maxPowerLimit.Int64),
				Memory:         uint64(memory.Int64),
				MinPowerLimit:  uint64(minPowerLimit.Int64),
				Model:          model,
				PowerLimit:     uint64(powerLimit.Int64),
				UUID:           uuid,
			},
		}
	}

	return querySlice[repr.SysinfoCardData](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadSacctData(
	filter types.DataProviderFilter,
) (recordBlobs [][]*repr.SacctInfo, softErrors int, err error) {
	var (
		account, allocTRES, cluster, distribution, jobStep, jobName       string
		jobState, partition, reservation, userName                        string
		nodes                                                             []string
		aveCPU, aveDiskRead, aveDiskWrite, aveRSS, aveVMSize, elapsedRaw  pgtype.Int8
		hetJobId, jobId, maxRSS, maxVMSize, minCPU, priority, suspendTime pgtype.Int8
		systemCPU, timeLimit, userCPU, arrayJobId                         pgtype.Int8
		hetJobOffset, requestedCpus, requestedMemoryPerNode               int
		requestedNodeCount                                                int
		endTime, startTime                                                pgtype.Timestamptz
		submitTime, timestamp                                             time.Time

		// Nullable, ignore NULL and translate to empty string or zero
		allocatedResourcesp, requestedResourcesp *string
		arrayTaskIdp, exitCodep                  *int
	)

	// The node -> nodes triggers special behavior in the query engine, too.
	q := query{
		DataProviderFilter: filter,
		fieldMap:           map[string]string{"time": "time", "user": "user_name", "job": "job_name", "node": "nodes"},
		table:              "sample_slurm_job",
		join: "join sample_slurm_job_acc as t2 on " +
			"t1.cluster = t2.cluster and t1.job_id = t2.job_id and t1.job_step = t2.job_step and " +
			"t1.time = t2.time",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "account, allocated_resources, \"AllocTRES\", array_job_id, array_task_id, \"AveCPU\", " +
			"\"AveDiskRead\", \"AveDiskWrite\", \"AveRSS\", \"AveVMSize\", t1.cluster, distribution, \"ElapsedRaw\", " +
			"end_time, exit_code, het_job_id, het_job_offset, t1.job_id, job_name, " +
			"job_state, t1.job_step, \"MaxRSS\", \"MaxVMSize\", \"MinCPU\", nodes, " +
			"partition, priority, requested_cpus, requested_memory_per_node, requested_node_count, " +
			"requested_resources, reservation, start_time, submit_time, suspend_time, \"SystemCPU\", " +
			"t1.time, time_limit, \"UserCPU\", user_name",
		boxes: []any{
			&account, &allocatedResourcesp, &allocTRES, &arrayJobId, &arrayTaskIdp, &aveCPU,
			&aveDiskRead, &aveDiskWrite, &aveRSS, &aveVMSize, &cluster, &distribution, &elapsedRaw,
			&endTime, &exitCodep, &hetJobId, &hetJobOffset, &jobId, &jobName,
			&jobState, &jobStep, &maxRSS, &maxVMSize, &minCPU, &nodes,
			&partition, &priority, &requestedCpus, &requestedMemoryPerNode, &requestedNodeCount,
			&requestedResourcesp, &reservation, &startTime, &submitTime, &suspendTime, &systemCPU,
			&timestamp, &timeLimit, &userCPU, &userName,
		},
	}

	// Reference: ParseSlurmV0JSON
	unbox := func() *repr.SacctInfo {
		var start, end int64
		if startTime.Valid {
			start = startTime.Time.UTC().Unix()
		}
		if endTime.Valid {
			end = endTime.Time.UTC().Unix()
		}
		var ajob uint32
		if arrayJobId.Valid {
			ajob = uint32(arrayJobId.Int64)
		}
		var allocatedResources, requestedResources string
		if allocatedResourcesp != nil {
			allocatedResources = *allocatedResourcesp
		}
		if requestedResourcesp != nil {
			requestedResources = *requestedResourcesp
		}
		var arrayTaskId, exitCode int
		if arrayTaskIdp != nil {
			arrayTaskId = *arrayTaskIdp
		}
		if exitCodep != nil {
			exitCode = *exitCodep
		}

		// The sonar version is currently lost in the timescaledb
		v0 := StringToUstr("0.0.0")

		// ArrayStep and HetStep appear to be lost and are UstrEmpty

		// For ReqGPUS, take whichever nonempty string comes first of allocatedResources and
		// requestedResources.  The former is more accurate, as the latter may have the fields in
		// some other order, but the former is not available for pending jobs.
		res := allocatedResources
		if res == "" {
			res = requestedResources
		}
		reqGpu := ""
		if res != "" {
			for _, f := range strings.Split(res, ",") {
				if strings.HasPrefix(f, "gres/gpu") {
					if reqGpu != "" {
						reqGpu = reqGpu + ","
					}
					reqGpu = reqGpu + f
				}
			}
		}

		nodeNames := ""
		for _, n := range nodes {
			// https://github.com/NordicHPC/sonar/issues/471 - Sonar should have scrubbed this
			// pointless output from sacct, instead it gets ingested as-is and stored in the
			// database.  So work around it.
			if n == "None assigned" {
				continue
			}
			if nodeNames != "" {
				nodeNames = nodeNames + ","
			}
			nodeNames = nodeNames + n
		}

		return &repr.SacctInfo{
			Time:         timestamp.UTC().Unix(),
			Start:        start,
			End:          end,
			Submit:       submitTime.UTC().Unix(),
			SystemCPU:    uint64(systemCPU.Int64),
			UserCPU:      uint64(userCPU.Int64),
			AveCPU:       uint64(aveCPU.Int64),
			MinCPU:       uint64(minCPU.Int64),
			Version:      v0,
			User:         StringToUstr(userName),
			JobName:      StringToUstr(jobName),
			State:        StringToUstr(jobState),
			Account:      StringToUstr(account),
			Layout:       StringToUstr(distribution),
			Reservation:  StringToUstr(reservation),
			JobStep:      StringToUstr(jobStep),
			NodeList:     StringToUstr(nodeNames),
			Partition:    StringToUstr(partition),
			ReqGPUS:      StringToUstr(reqGpu),
			AllocRes:     StringToUstr(allocatedResources),
			ReqRes:       StringToUstr(requestedResources),
			JobID:        uint32(jobId.Int64),
			ArrayJobID:   ajob,
			ArrayTaskID:  uint32(arrayTaskId),
			HetJobID:     uint32(hetJobId.Int64),
			HetJobOffset: uint32(hetJobOffset),
			AveDiskRead:  uint64(aveDiskRead.Int64),
			AveDiskWrite: uint64(aveDiskWrite.Int64),
			AveRSS:       uint64(aveRSS.Int64),
			AveVMSize:    uint64(aveVMSize.Int64),
			ElapsedRaw:   uint32(elapsedRaw.Int64),
			MaxRSS:       uint64(maxRSS.Int64),
			MaxVMSize:    uint64(maxVMSize.Int64),
			ReqCPUS:      uint32(requestedCpus),
			ReqMem:       uint64(requestedMemoryPerNode),
			ReqNodes:     uint32(requestedNodeCount),
			Suspended:    uint32(suspendTime.Int64),
			TimelimitRaw: uint32(timeLimit.Int64),
			ExitCode:     uint8(exitCode),
			Priority:     uint64(priority.Int64),
		}
	}

	return querySlice[repr.SacctInfo](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadCluzterAttributeData(
	filter types.DataProviderFilter,
) (recordBlobs [][]*repr.CluzterAttributes, softErrors int, err error) {
	var (
		cluster   string
		slurm     bool
		timestamp time.Time
	)

	q := query{
		DataProviderFilter: filter,
		table:              "cluster_attributes",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "cluster, slurm, time",
		boxes:  []any{&cluster, &slurm, &timestamp},
	}

	// Reference: ParseCluzterV0JSON
	unbox := func() *repr.CluzterAttributes {
		return &repr.CluzterAttributes{
			Time:    timestamp.Format(time.RFC3339),
			Cluster: cluster,
			Slurm:   slurm,
		}
	}

	return querySlice[repr.CluzterAttributes](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadCluzterPartitionData(
	filter types.DataProviderFilter,
) (recordBlobs [][]*repr.CluzterPartitions, softErrors int, err error) {
	var (
		cluster          string
		partName         string
		nodeNamesCompact []string
		timestamp        time.Time
	)

	q := query{
		DataProviderFilter: filter,
		table:              "partition",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "cluster, nodes_compact, partition, time",
		boxes:  []any{&cluster, &nodeNamesCompact, &partName, &timestamp},
	}

	// A little tricky.  The Sonar data has multiple partitions for a timestamp in the same object
	// (in fact all partitions on the cluster at that time).  The database has flattened this view
	// and each record has only one partition.  The only consumer of partition data at this point
	// breaks down the Sonar view immediately, so it's fine not to merge anything here.

	// Reference: ParseCluzterV0JSON
	unbox := func() *repr.CluzterPartitions {
		nodes := make([]newfmt.HostnameRange, 0, len(nodeNamesCompact))
		for _, nnc := range nodeNamesCompact {
			nodes = append(nodes, newfmt.HostnameRange(nnc))
		}
		return &repr.CluzterPartitions{
			Time:    timestamp.Format(time.RFC3339),
			Cluster: cluster,
			Partitions: []newfmt.ClusterPartition{
				newfmt.ClusterPartition{
					Name:  newfmt.NonemptyString(partName),
					Nodes: nodes,
				},
			},
		}
	}

	return querySlice[repr.CluzterPartitions](cdb, &q, unbox)
}

func (cdb *connectedDB) ReadCluzterNodeData(
	filter types.DataProviderFilter,
) (recordBlobs [][]*repr.CluzterNodes, softErrors int, err error) {
	var (
		cluster   string
		nodeName  string
		states    []string
		timestamp time.Time
	)

	q := query{
		DataProviderFilter: filter,
		table:              "node_state",
		// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
		fields: "cluster, node, states, time",
		boxes:  []any{&cluster, &nodeName, &states, &timestamp},
	}

	// TODO: Merge nodes in the same state, as the consumer (snodes) depends on that for good UX.
	// It could also be fixed in the consumer...

	// Reference: ParseCluzterV0JSON
	unbox := func() *repr.CluzterNodes {
		return &repr.CluzterNodes{
			Time:    timestamp.Format(time.RFC3339),
			Cluster: cluster,
			Nodes: []newfmt.ClusterNodes{
				newfmt.ClusterNodes{
					Names:  []newfmt.HostnameRange{newfmt.HostnameRange(nodeName)},
					States: states,
				},
			},
		}
	}

	return querySlice[repr.CluzterNodes](cdb, &q, unbox)
}

func (cdb *connectedDB) MinTime(soft bool) (time.Time, error) {
	return cdb.timeCache.MinTime(soft)
}

func (cdb *connectedDB) MaxTime(soft bool) (time.Time, error) {
	return cdb.timeCache.MaxTime(soft)
}

func makeRefillTimeCache(
	dbc *databaseConnection,
	clusterName string,
) func() (time.Time, time.Time, error) {
	return func() (low, high time.Time, err error) {
		err = dbc.QueryRowAndScan(
			context.Background(),
			"SELECT MIN(time) FROM sysinfo_attributes WHERE cluster = $1",
			[]any{clusterName},
			[]any{&low},
		)
		if err != nil {
			return
		}
		err = dbc.QueryRowAndScan(
			context.Background(),
			"SELECT MAX(time) FROM sample_process WHERE cluster = $1",
			[]any{clusterName},
			[]any{&high},
		)
		return
	}
}

// Totally gross.  Keep this in sync with logic in querySlice!!
func toDateName(fromDate, toDate time.Time) string {
	n := 2
	if !fromDate.IsZero() {
		n++
	}
	if toDate.IsZero() {
		panic("Should not happen")
	}
	return fmt.Sprintf("$%d", n)
}

func querySlice[T any](
	cdb *connectedDB,
	q *query,
	unbox func() *T,
) (finalRows [][]*T, softErrors int, err error) {
	primary := "SELECT * FROM " + q.table + " WHERE " + "cluster=$1"
	qarg := []any{cdb.cx.ClusterName()}

	// A note about adding field filters against values.  For values we don't control we have to use
	// parameters to avoid being pwned by Mrs Roberts.  For values that we do control (typically b/c
	// they are numbers or dates) we can expand the values inline.
	//
	// It might be better for query optimization / compilation / reuse to make also the in-line
	// values parameters, but I don't know this yet.

	// Keep these in sync with toDateName above!!  Note that even though dates could be in-line
	// values they are parameters here to allow reuse by other parts of the query.  Sort of hacky.

	timeField := mapField("time", q.fieldMap)
	if !q.FromDate.IsZero() {
		primary += fmt.Sprintf(" AND %s >= $%d", timeField, len(qarg)+1)
		qarg = append(qarg, q.FromDate.Format(time.DateOnly))
	}
	if !q.ToDate.IsZero() {
		primary += fmt.Sprintf(" AND %s < $%d", timeField, len(qarg)+1)
		qarg = append(qarg, q.ToDate.Add(time.Hour*24).Format(time.DateOnly))
	}

	// Add host filters.
	//
	// At the database level, host filtering is exclusively an optimization, to avoid reading /
	// generating data.  Note in particular that we can apply lossy abbreviations as long as they
	// find everything a precise match would find.

	if q.Node != nil && !q.Node.IsEmpty() {
		fieldname := mapField("node", q.fieldMap)
		if fieldname == "nodes" {
			// As a special hack, when "node" maps to "nodes" then the selector turns into set
			// intersection, and the queried node set is expanded into an array that is intersected
			// with an array field called "nodes".  This is gross but probably OK for now.  The
			// meaning is always that the sets overlap, not that the lhs is contained in the rhs.
			// This is debatable but since this is an optimization and post-filtering must happen
			// anyway it is probably the right thing.
			var x, expanded []string
			var err error
			for _, p := range q.Node.Patterns() {
				x, err = hostglob.ExpandPattern(p)
				if err != nil {
					break
				}
				expanded = append(expanded, x...)
			}
			if err == nil && len(expanded) > 0 && len(expanded) <= dosCutoff {
				elements := make([]string, 0, len(expanded))
				for _, e := range expanded {
					elements = append(elements, fmt.Sprintf("$%d::character varying", len(qarg)+1))
					qarg = append(qarg, e)
				}
				primary += " AND " + fieldname + " && ARRAY[" + strings.Join(elements, ",") + "]"
			}
		} else {
			conds := make([]string, 0)
			args := make([]any, 0)
			nextIx := len(qarg) + 1
			for _, p := range q.Node.Patterns() {
				loc := strings.IndexAny(p, "[*")
				if !q.Node.IsPrefix() && loc == -1 {
					conds = append(conds, fmt.Sprintf("%s = $%d", fieldname, nextIx))
				} else {
					// TODO: We can and should do more here:
					//
					// Some ranges can usefully be expanded into prefixes but it can be tricky:
					// c1-[10-15] becomes c1-1 but c1-[9-20] becomes c1-1 OR c1-2 OR c1-9, among other
					// things.  Yet simple cases of this are important, as e.g. gpu-[8,9] is a common
					// thing.
					//
					// We could do 'c1-*.fox' as 'like c1-%.fox' and it may be worthwhile to do so,
					// ditto c1-[10-40].fox could be 'like c1-%.fox'.  We could do c1-[2-8] as 'like
					// 'c1-_' and it might be worthwhile (% matches zero or more, _ matches one).
					//
					// A perf hint for % is to avoid leading wildcards.
					//
					// We should get the hostglobber involved since it has an exact parse of the pattern
					// set and is already in q.hosts.
					conds = append(conds, fmt.Sprintf("%s like $%d", fieldname, nextIx))
					if loc != -1 {
						p = p[:loc]
					}
					p += "%"
				}
				args = append(args, p)
				nextIx++
			}
			if len(conds) <= dosCutoff {
				primary += " AND (" + strings.Join(conds, " OR ") + ")"
				qarg = append(qarg, args...)
			}
		}
	}

	// Add job id filters.

	if len(q.Jobs) > 0 && len(q.Jobs) <= dosCutoff {
		fieldname := mapField("job", q.fieldMap)
		var jobs []string
		for j := range q.Jobs {
			jobs = append(jobs, fmt.Sprintf("%s = %d", fieldname, j))
		}
		primary += " AND (" + strings.Join(jobs, " OR ") + ")"
	}

	// Add user name filters.

	if len(q.Users) > 0 && len(q.Users) <= dosCutoff {
		fieldname := mapField("user", q.fieldMap)
		var conds []string
		for u := range q.Users {
			conds = append(conds, fmt.Sprintf("%s = $%d", fieldname, len(qarg)+1))
			qarg = append(qarg, u.String())
		}
		primary += " AND (" + strings.Join(conds, " OR ") + ")"
	}

	qstr := "SELECT " + q.fields + " FROM (" + primary + ") AS t1"
	if q.join != "" {
		qstr += " " + q.join
	}
	if Verbose {
		Log.Infof("SQL: %s %s", qstr, qarg)
	}
	rows, err := cdb.theDB.Query(context.Background(), qstr, qarg...)
	if err != nil {
		return
	}
	dataRows := make([]*T, 0)
	_, err = pgx.ForEachRow(rows, q.boxes, func() error {
		dataRows = append(dataRows, unbox())
		return nil
	})
	if err != nil {
		if Verbose {
			Log.Warningf("SQL: %v", err)
		}
		return
	}
	if Verbose {
		Log.Infof("SQL: Retrieved %d rows", len(dataRows))
	}
	finalRows = [][]*T{dataRows}
	return
}

func mapField(field string, fieldMap map[string]string) string {
	if fieldMap == nil {
		return field
	}
	probe := fieldMap[field]
	if probe == "" {
		panic("Unmapped field name in SQL query: " + field)
	}
	return probe
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
