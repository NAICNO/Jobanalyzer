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
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	. "sonalyze/common"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

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

func OpenDatabaseURI(databaseURI string) (*databaseConnection, error) {
	connection, err := pgx.Connect(context.Background(), databaseURI)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v\n", err)
	}
	return &databaseConnection{connection: connection}, nil
}

func (cdb *databaseConnection) EnumerateClusters() ([]string, error) {
	rows, err := cdb.Query(context.Background(), "SELECT cluster FROM cluster_attributes")
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

	table := "sample_process"
	clusterTable := ""

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
			// TODO: Gpus, GpuPct, GpuMemPct, GpuFail - requires some kind of join probably with
			// sample_process_gpu although the index is very complicated, five fields here and six
			// fields there.  And we'll find multiple GPU records per sample record, which will be a
			// mess.
			//
			// TODO: Cores - ditto, come from sysinfo_attributes, join on time (approximate),
			// cluster, node probably.
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
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, table, fields, clusterTable)
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

	table := "sample_system"

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
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, table, fields, "")
}

func (cdb *connectedDB) ReadCpuSamples(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (dataBlobs [][]*repr.CpuSamples, softErrors int, err error) {
	var cpus []pgtype.Int8
	var node string
	var timestamp time.Time

	table := "sample_system"

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
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, table, fields, "")
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

	table := "sysinfo_attributes"

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
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, table, fields, "")
}

func (cdb *connectedDB) ReadSysinfoCardData(
	fromDate, toDate time.Time,
	hosts *Hosts,
	verbose bool,
) (sysinfoBlobs [][]*repr.SysinfoCardData, softErrors int, err error) {
	var address, architecture, driver, firmware, manufacturer, node, uuid string
	var index int
	var maxCeClock, maxMemoryClock, maxPowerLimit, memory, minPowerLimit, powerLimit pgtype.Int8
	var timestamp time.Time

	// The DB stores what it perceives to be static card info in a separate table, sysinfo_gpu_card.
	// That needs to be joined to sysinfo_gpu_card_config here (by UUID) to get the full story.
	table := "sysinfo_gpu_card_config t1 join sysinfo_gpu_card t2 on t1.uuid = t2.uuid"

	// Alpha field name order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "address, architecture, driver, firmware, index, manufacturer, max_ce_clock, max_memory_clock, " +
		"max_power_limit, memory, min_power_limit, node, power_limit, time, t1.uuid"
	boxes := []any{
		&address, &architecture, &driver, &firmware, &index, &manufacturer, &maxCeClock, &maxMemoryClock,
		&maxPowerLimit, &memory, &minPowerLimit, &node, &powerLimit, &timestamp, &uuid,
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
				MaxCEClock:     uint64(maxCeClock.Int),
				MaxMemoryClock: uint64(maxMemoryClock.Int),
				MaxPowerLimit:  uint64(maxPowerLimit.Int),
				Memory:         uint64(memory.Int),
				MinPowerLimit:  uint64(minPowerLimit.Int),
				PowerLimit:     uint64(powerLimit.Int),
				UUID:           uuid,
			},
		}
	}

	return querySlice[repr.SysinfoCardData](
		cdb, fromDate, toDate, hosts, verbose, boxes, unbox, table, fields, "")
}

func (cdb *connectedDB) ReadSacctData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.SacctInfo, softErrors int, err error) {
	var (
		account, allocTRES, cluster, distribution, jobStep, jobName         string
		jobState, partition, reservation, userName                          string
		allocatedResources, requestedResources                              string
		nodes                                                               []string
		aveCPU, aveDiskRead, aveDiskWrite, aveRSS, aveVMSize, elapsedRaw    pgtype.Int8
		hetJobId, jobId, maxRSS, maxVMSize, minCPU, priority, suspendTime   pgtype.Int8
		systemCPU, timeLimit, userCPU, arrayJobId                           pgtype.Int8
		arrayTaskId, exitCode                                               *int
		minCpusPerNode, hetJobOffset, requestedCpus, requestedMemoryPerNode int
		requestedNodeCount                                                  int
		endTime, startTime                                                  pgtype.Timestamptz
		submitTime, timestamp                                               time.Time
	)

	table := "sample_slurm_job as t1 join sample_slurm_job_acc as t2 on " +
		"t1.cluster = t2.cluster and " +
		"t1.job_id = t2.job_id and " +
		"t1.job_step = t2.job_step and " +
		"t1.time = t2.time"

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "account, allocated_resources, \"AllocTRES\", array_job_id, array_task_id, \"AveCPU\", " +
		"\"AveDiskRead\", \"AveDiskWrite\", \"AveRSS\", \"AveVMSize\", t1.cluster, distribution, \"ElapsedRaw\", " +
		"end_time, exit_code, het_job_id, het_job_offset, t1.job_id, job_name, " +
		"job_state, t1.job_step, \"MaxRSS\", \"MaxVMSize\", \"MinCPU\", minimum_cpus_per_node, nodes, " +
		"partition, priority, requested_cpus, requested_memory_per_node, requested_node_count, " +
		"requested_resouces, reservation, start_time, submit_time, suspend_time, \"SystemCPU\", " +
		"t1.time, time_limit, \"UserCPU\", user_name"
	boxes := []any{
		&account, &allocatedResources, &allocTRES, &arrayJobId, &arrayTaskId, &aveCPU,
		&aveDiskRead, &aveDiskWrite, &aveRSS, &aveVMSize, &cluster, &distribution, &elapsedRaw,
		&endTime, &exitCode, &hetJobId, &hetJobOffset, &jobId, &jobName,
		&jobState, &jobStep, &maxRSS, &maxVMSize, &minCPU, &minCpusPerNode, &nodes,
		&partition, &priority, &requestedCpus, &requestedMemoryPerNode, &requestedNodeCount,
		&requestedResources, &reservation, &startTime, &submitTime, &suspendTime, &systemCPU,
		&timestamp, &timeLimit, &userCPU, &userName,
	}

	// Reference: ParseSlurmV0JSON
	unbox := func() *repr.SacctInfo {
		// Handle nullable fields
		var start, end int64
		if startTime.Status == pgtype.Present {
			start = startTime.Time.UTC().Unix()
		}
		if endTime.Status == pgtype.Present {
			end = endTime.Time.UTC().Unix()
		}
		var ajob, atask, xcode uint32
		if arrayJobId.Status == pgtype.Present {
			ajob = uint32(arrayJobId.Int)
		}
		if arrayTaskId != nil {
			atask = uint32(*arrayTaskId)
		}
		if exitCode != nil {
			xcode = uint32(*exitCode)
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
			SystemCPU:    uint64(systemCPU.Int),
			UserCPU:      uint64(userCPU.Int),
			AveCPU:       uint64(aveCPU.Int),
			MinCPU:       uint64(minCPU.Int),
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
			JobID:        uint32(jobId.Int),
			ArrayJobID:   ajob,
			ArrayIndex:   atask,
			HetJobID:     uint32(hetJobId.Int),
			HetOffset:    uint32(hetJobOffset),
			AveDiskRead:  uint32(aveDiskRead.Int),
			AveDiskWrite: uint32(aveDiskWrite.Int),
			AveRSS:       uint32(aveRSS.Int),
			AveVMSize:    uint32(aveVMSize.Int),
			ElapsedRaw:   uint32(elapsedRaw.Int),
			MaxRSS:       uint32(maxRSS.Int),
			MaxVMSize:    uint32(maxVMSize.Int),
			ReqCPUS:      uint32(requestedCpus),
			ReqMem:       uint32(requestedMemoryPerNode),
			ReqNodes:     uint32(requestedNodeCount),
			Suspended:    uint32(suspendTime.Int),
			TimelimitRaw: uint32(timeLimit.Int),
			ExitCode:     uint8(xcode),
		}
	}

	return querySlice[repr.SacctInfo](
		cdb, fromDate, toDate, nil, verbose, boxes, unbox, table, fields, "t1.")
}

func (cdb *connectedDB) ReadCluzterAttributeData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterAttributes, softErrors int, err error) {
	var cluster string
	var slurm bool
	var timestamp time.Time

	table := "cluster_attributes"

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "cluster, slurm, time"
	boxes := []any{&cluster, &slurm, &timestamp}

	// Reference: ParseCluzterV0JSON
	unbox := func() *repr.CluzterAttributes {
		return &repr.CluzterAttributes{
			Time:    timestamp.Format(time.RFC3339),
			Cluster: cluster,
			Slurm:   slurm,
		}
	}

	return querySlice[repr.CluzterAttributes](
		cdb, fromDate, toDate, nil, verbose, boxes, unbox, table, fields, "")
}

func (cdb *connectedDB) ReadCluzterPartitionData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterPartitions, softErrors int, err error) {
	var cluster string
	var partName string
	var nodeNamesCompact []string
	var timestamp time.Time

	table := "partition"

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "cluster, nodes_compact, partition, time"
	boxes := []any{&cluster, &nodeNamesCompact, &partName, &timestamp}

	// A little tricky.  The Sonar data has multiple partitions for a timestamp in the same object
	// (in fact all partitions on the cluster at that time).  The database has flattened this view
	// and each record has only one partition.  The only consumer of partition data at this point
	// breaks down the Sonar view immediately, so it's fine not to merge anything here.

	// Reference: ParseCluzterV0JSON
	unbox := func() *repr.CluzterPartitions {
		nodes := make([]newfmt.NodeRange, 0, len(nodeNamesCompact))
		for _, nnc := range nodeNamesCompact {
			nodes = append(nodes, newfmt.NodeRange(nnc))
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

	return querySlice[repr.CluzterPartitions](
		cdb, fromDate, toDate, nil, verbose, boxes, unbox, table, fields, "")
}

func (cdb *connectedDB) ReadCluzterNodeData(
	fromDate, toDate time.Time,
	verbose bool,
) (recordBlobs [][]*repr.CluzterNodes, softErrors int, err error) {
	var cluster string
	var nodeName string
	var states []string
	var timestamp time.Time

	table := "node_state"

	// Alpha order and KEEP THESE TWO LISTS COMPLETELY IN SYNC OR YOU WILL BE SORRY!
	fields := "cluster, node, states, time"
	boxes := []any{&cluster, &nodeName, &states, &timestamp}

	// TODO: Merge nodes in the same state, as the consumer (snodes) depends on that for good UX.
	// It could also be fixed in the consumer...

	// Reference: ParseCluzterV0JSON
	unbox := func() *repr.CluzterNodes {
		return &repr.CluzterNodes{
			Time:    timestamp.Format(time.RFC3339),
			Cluster: cluster,
			Nodes: []newfmt.ClusterNodes{
				newfmt.ClusterNodes{
					Names:  []newfmt.NodeRange{newfmt.NodeRange(nodeName)},
					States: states,
				},
			},
		}
	}

	return querySlice[repr.CluzterNodes](
		cdb, fromDate, toDate, nil, verbose, boxes, unbox, table, fields, "")
}

func querySlice[T any](
	cdb *connectedDB,
	fromDate, toDate time.Time,
	hosts *Hosts, // may be nil
	verbose bool,
	boxes []any,
	unbox func() *T,
	table, fields string,
	clusterPrefix string,
) (finalRows [][]*T, softErrors int, err error) {
	// TODO: time span and hosts obviously
	// TODO: precedence of 'table' (which may be a join) vis-a-vis the WHERE clause
	qstr := "SELECT " + fields + " FROM " + table + " WHERE " + clusterPrefix + "cluster=$1"
	qarg := []any{cdb.cx.ClusterName()}
	rows, err := cdb.theDB.Query(context.Background(), qstr, qarg...)
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
