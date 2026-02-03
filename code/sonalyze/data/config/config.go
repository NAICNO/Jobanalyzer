// Query static configuration data, ie, the join of node and card data with the option of some
// filtering.  This is a computed view.  Computed values are held in memory, but never purged.
package config

import (
	"fmt"
	"math"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	umaps "go-utils/maps"

	. "sonalyze/common"
	"sonalyze/data/card"
	"sonalyze/data/common"
	"sonalyze/data/node"
	"sonalyze/db/repr"
	"sonalyze/db/types"
)

type NodeConfig struct {
	repr.NodeSummary
	Time      int64
	Distances string
	TopoSVG   string
	TopoText  string
}

type ConfigDataProvider struct {
	meta  types.Context
	data  *perClusterInfo
	valid bool
	cards *card.CardDataProvider
	nodes *node.NodeDataProvider
}

var (
	// MT: Atomic
	// The number of live cache entries - not currently used for much
	globalOccupancy atomic.Uint64

	// MT: Locked
	//
	// Maps cluster name to unique per-cluster data.  We could have attached that to the
	// ClusterEntry but for now it's nice to keep it local to this package.
	clusterTableLock sync.Mutex
	clusterTable     = make(map[string]*perClusterInfo)
)

func OpenConfigDataProvider(meta types.Context) (*ConfigDataProvider, error) {
	nodes, err := node.OpenNodeDataProvider(meta)
	if err != nil {
		return nil, err
	}
	cards, err := card.OpenCardDataProvider(meta)
	if err != nil {
		return nil, err
	}

	name := meta.ClusterName()

	clusterTableLock.Lock()
	defer clusterTableLock.Unlock()

	data := clusterTable[name]
	if data == nil {
		data = makePerClusterInfo(name)
		clusterTable[name] = data
	}

	return &ConfigDataProvider{
		meta:  meta,
		data:  data,
		valid: true,
		nodes: nodes,
		cards: cards,
	}, nil
}

func MaybeOpenConfigDataProvider(meta types.Context) *ConfigDataProvider {
	cdp, err := OpenConfigDataProvider(meta)
	if err == nil {
		return cdp
	}
	return &ConfigDataProvider{
		meta: meta,
	}
}

// This can return nil.  We want the latest host information at or before the given time, which is
// seconds since Unix epoch UTC.  If the database has to be queried, the query window into the past
// may be limited to 14 days.  The result is not necessarily stable, it may change if new data come
// in, but will never revert to older data.  New data that replace a prior non-nil result may or may
// not be honored in a timely manner.  A static cluster configuration, should it exist, will be
// consulted only if the information can't be found in the database.
func (cdp *ConfigDataProvider) LookupHostByTime(host Ustr, t int64) *repr.NodeSummary {
	if !cdp.valid {
		return nil
	}

	err := cdp.populateCache(QueryArgs{
		QueryFilter: common.QueryFilter{
			HaveFrom: true,
			FromDate: time.Unix(t-(60*60*24*14), 0).UTC(),
			HaveTo:   true,
			ToDate:   time.Unix(t, 0).UTC(),
			Host:     []string{host.String()}, // groan!!!
		},
	})
	if err == nil {
		if tmp := cdp.obtainOneFromCache(host.String(), t); tmp != nil {
			return &tmp.NodeSummary
		}
	}

	// Fallback code to old-style static node config, this will likely disappear.
	if cdp.meta.HaveConfig() {
		return cdp.meta.Config().LookupHost(host.String())
	}

	return nil
}

// This is a primitive API that provides a set of all nodes that are present in the data in the
// given time interval, or in the backing config file if there is one and there are no nodes in the
// data base.
//
// Noe this cannot be taken from computed config data because need this list to compute the config
// data for an open set of hosts.
//
// This needs a cache and/or lazy computation, too.  The right way to think about it, I think, is to
// ignore the toDate, and to populate the lazy table from the fromDate to the youngest date not
// covered, which could be today's date.  The table should have one entry per calendar day (this is
// good enough) and should be careful to share data where possible, because the host sets can be
// quite large but have tremendous stability over time, so there will be a lot of sharing.
//
// This API can be made internal, the only external user can use ConfigDataProvider.Query for the
// same effect.  And it's not necessary for the return value to be a map, both current consumers
// really want a list of names, neither uses the map as a map.  But maps.Equal runs in O(n) time, so
// maps are no worse than slices locally.

func (cdb *ConfigDataProvider) AvailableHosts(fromDate, toDate time.Time) (map[string]bool, error) {
	recordBlobs, _, err := cdb.nodes.QueryRaw(
		fromDate,
		toDate,
		nil,
		false,
	)
	if err != nil {
		return nil, err
	}
	nodenames := make(map[string]bool)
	if len(recordBlobs) > 0 {
		for _, blob := range recordBlobs {
			for _, record := range blob {
				nodenames[record.Node] = true
			}
		}
	} else {
		for _, node := range cdb.meta.NodesDefinedInConfigIfAny() {
			nodenames[node.Hostname] = true
		}
	}
	return nodenames, nil
}

type QueryArgs struct {
	common.QueryFilter
	Verbose bool
	Newest  bool
	Query   func(records []*NodeConfig) ([]*NodeConfig, error)
}

// Note clients will need to set the query parameters sensibly.  Before, when we had a single
// time-invariant config file, that was not necessary, but now it is.
//
// As for all other query operators, if the host set is empty (the common case) then we find all
// hosts that have data in the time range.

func (cdp *ConfigDataProvider) Query(qa QueryArgs) ([]*NodeConfig, error) {
	if len(qa.Host) == 0 {
		hosts, err := cdp.AvailableHosts(qa.FromDate, qa.ToDate)
		if err != nil {
			return nil, err
		}
		qa.Host = umaps.Keys(hosts)
	}

	if !cdp.valid || len(qa.Host) == 0 {
		return make([]*NodeConfig, 0), nil
	}

	err := cdp.populateCache(qa)
	if err != nil {
		return nil, err
	}

	records := cdp.obtainAllFromCache(qa)

	// Fallback code to old-style static node config, this will likely disappear.
	if len(records) == 0 && cdp.meta.HaveConfig() {
		for _, host := range qa.Host {
			if probe := cdp.meta.Config().LookupHost(host); probe != nil {
				records = append(records, &NodeConfig{
					NodeSummary: *probe,
				})
			}
		}
	}

	if qa.Query != nil {
		records, err = qa.Query(records)
		if err != nil {
			return nil, err
		}
	}
	if qa.Newest {
		newr := make(map[string]*NodeConfig)
		for _, r := range records {
			if probe := newr[r.Hostname]; probe != nil {
				if r.Timestamp > probe.Timestamp {
					newr[r.Hostname] = r
				}
			} else {
				newr[r.Hostname] = r
			}
		}
		records = umaps.Values(newr)
	}
	return records, nil
}

// Raw query against the database, synthesizing records from node and card data and running
// independent of any caching.  It's fine for there to be multiple hosts in the query args, but the
// result should be independent of whether a single materialize() is run on multiple hosts or one
// materialize() is run for each host.

func (cdp *ConfigDataProvider) materialize(qa QueryArgs) ([]*NodeConfig, error) {
	nodes, err := cdp.nodes.Query(
		node.QueryFilter{
			HaveFrom: qa.HaveFrom,
			FromDate: qa.FromDate,
			HaveTo:   qa.HaveTo,
			ToDate:   qa.ToDate,
			Host:     qa.Host,
		},
		qa.Verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	cards, err := cdp.cards.Query(
		card.QueryFilter{
			HaveFrom: qa.HaveFrom,
			FromDate: qa.FromDate,
			HaveTo:   qa.HaveTo,
			ToDate:   qa.ToDate,
			Host:     qa.Host,
		},
		qa.Verbose,
	)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log records: %v", err)
	}
	type joinedData struct {
		// the time and host are given by node
		node  *repr.SysinfoNodeData
		cards []*repr.SysinfoCardData
	}
	joined := make(map[string]*joinedData)
	for _, r := range nodes {
		joined[r.Time+"|"+r.Node] = &joinedData{node: r}
	}
	for _, r := range cards {
		if probe := joined[r.Time+"|"+r.Node]; probe != nil {
			probe.cards = append(probe.cards, r)
		}
	}
	rawRecords := umaps.Values(joined)
	records := make([]*NodeConfig, len(rawRecords))
	for i, r := range rawRecords {
		ht := ""
		if r.node.ThreadsPerCore > 1 {
			ht = " (hyperthreaded)"
		}
		memGB := int(math.Round(float64(r.node.Memory) / (1024 * 1024)))
		desc := fmt.Sprintf(
			"%dx%d%s %s, %d GiB", r.node.Sockets, r.node.CoresPerSocket, ht, r.node.CpuModel, memGB)
		cores := r.node.Sockets * r.node.CoresPerSocket * r.node.ThreadsPerCore
		numCards := len(r.cards)
		cardTotalMemKB := uint64(0)
		for _, c := range r.cards {
			cardTotalMemKB += c.Memory
		}
		cardTotalMemGB := int(math.Round(float64(cardTotalMemKB) / (1024 * 1024)))
		if numCards > 0 {
			desc += fmt.Sprintf(", %dx %s @ %dGiB", numCards, r.cards[0].Model, (r.cards[0].Memory)/(1024*1024))
		}
		distances := ""
		if r.node.Distances != nil {
			distances = fmt.Sprintf("%v", r.node.Distances)
		}
		var parsedTime int64
		parsedTimeTmp, err := time.Parse(time.RFC3339, r.node.Time)
		if err == nil {
			parsedTime = parsedTimeTmp.UTC().Unix()
		}
		records[i] = &NodeConfig{
			NodeSummary: repr.NodeSummary{
				Timestamp:   r.node.Time,
				Hostname:    r.node.Node,
				Description: desc,
				CpuCores:    int(cores),
				MemGB:       memGB,
				GpuCards:    numCards,
				GpuMemGB:    cardTotalMemGB,
				// CrossNodeJobs is not being set here because it is ill-defined and will
				// be removed from sonalyze.
				//
				// `Metadata` is unused by sonalyze.
			},
			Time:      parsedTime,
			Distances: distances,
			TopoSVG:   r.node.TopoSVG,
			TopoText:  r.node.TopoText,
		}
	}
	return records, nil
}

// NodeConfig cache:
//
// - For every a host, we have a table of records (arranged in such a way that we can perform
//   time-ranged queries on it and insert at both ends).  Additionally there is the oldest timestamp
//   we have scanned for (which may be older than the oldest record), the timestamps of the oldest
//   and youngest records, and the time of the last young-record scan.
//
// - If we're requesting a record that is older than the oldest timestamp we have scanned for then,
//   we must scan the database for earlier times and we'll scan from the time we're looking for
//   minus 2 weeks forward to the oldest time we've scanned for previously, and populate the table
//   with what we find, and set the oldest scan time appropriately.
//
// - If we're requesting a time that is younger than the youngest record and it has been more than 1
//   hour since we scanned for younger records, then scan for the range of the current time to the
//   youngest record we have, and populate the table.
//
// - Data that match by timestamp will replace existing data in the table with the same timestamp.
//
// After the scanning, we can just grab data from the cache.
//
// Note that per-host tables will tend to be short (because mostly we're interested in recent data).
// But the tables can become long (if very old data are requested).  We may see data out of order
// in the database so insertions in the middle somewhere is a possibility but mostly this will
// happen near the ends.
//
// NOTE: This is a very general idea, there is almost nothing here that is specific to NodeConfig
// data.  It could be parameterized by the type of data and a small method suite for examining data
// and materializing them.  It then becomes a general caching layer for computed views (although as
// usual invalidation is an issue).  The Time field that got added to NodeConfig can probably be
// attached externally, in the cache, we just need a means of computing it from the data.
//
// Implementation:
//
// For now, keep it very, very simple: One per-cluster lock over a linked structure held in a
// per-cluster variable.  There will be contention, but it's a lazily built data structure and
// eventually everything will settle down.  We can optimize later.
//
// per-cluster-info : host-name -> per-host-info
// per-host-info : always-sorted timestamp-unique list of records + metadata

type perClusterInfo struct {
	name          string
	hostTableLock sync.Mutex
	hostTable     map[string]*hostInfo
}

func makePerClusterInfo(name string) *perClusterInfo {
	return &perClusterInfo{
		name:      name,
		hostTable: make(map[string]*hostInfo),
	}
}

type hostInfo struct {
	records           []*NodeConfig // ascending by time, no repeated timestamps
	lastScan          int64
	oldestScannedTime int64
	youngestRecord    int64
	oldestRecord      int64
}

func (hr *hostInfo) insertLocked(r *NodeConfig) (inserted bool) {
	if r.Time > hr.youngestRecord {
		hr.records = append(hr.records, r)
		inserted = true
	} else {
		// TODO: binary search probably
		i := 0
		for i < len(hr.records) && r.Time > hr.records[i].Time {
			i++
		}
		if r.Time != hr.records[i].Time {
			hr.records = slices.Insert(hr.records, i, r)
			inserted = true
		}
	}
	hr.youngestRecord = max(r.Time, hr.youngestRecord)
	hr.oldestRecord = min(r.Time, hr.oldestRecord)
	return
}

func (hr *hostInfo) getLocked(from int64, to int64) []*NodeConfig {
	low := 0
	for low < len(hr.records) && hr.records[low].Time < from {
		low++
	}
	high := len(hr.records) - 1
	for high >= low && hr.records[high].Time > to {
		high--
	}
	return hr.records[low : high+1]
}

func (cdp *ConfigDataProvider) populateCache(qa QueryArgs) error {
	// Split into computeWorklist and insert so that the materialization does not need to hold the
	// lock.  The work items will only ever pertain to one host, for now.

	type chunk struct {
		workItem QueryArgs
		records  []*NodeConfig
	}

	chunks := make([]chunk, 0)
	for _, workItem := range cdp.computeWorklist(qa) {
		records, err := cdp.materialize(workItem)
		if err != nil {
			return err
		}
		chunks = append(chunks, chunk{workItem, records})
	}

	perCluster := cdp.data
	perCluster.hostTableLock.Lock()
	defer perCluster.hostTableLock.Unlock()

	// Each chunk pertains to exactly one host (due to restriction on work item) and the host will
	// have an entry in the table.

	now := time.Now().UTC().Unix()
	for _, c := range chunks {
		hn := c.workItem.Host[0]
		perHost := perCluster.hostTable[hn]
		if perHost == nil {
			panic("Inconsistent table: no entry for host " + hn)
		}
		perHost.lastScan = now
		for _, r := range c.records {
			if perHost.insertLocked(r) {
				globalOccupancy.Add(1)
			}
		}
	}
	return nil
}

const (
	oneHour  = 60 * 60
	oneDay   = 60 * 14
	twoWeeks = oneDay * 14
)

func (cdp *ConfigDataProvider) computeWorklist(qa QueryArgs) []QueryArgs {
	perCluster := cdp.data
	perCluster.hostTableLock.Lock()
	defer perCluster.hostTableLock.Unlock()

	nowt := time.Now().UTC()
	now := nowt.Unix()
	var fromTime, toTime int64
	toTime = qa.ToDate.Unix()
	fromTime = qa.FromDate.Unix()
	worklist := make([]QueryArgs, 0)
	for _, hn := range qa.Host {
		perHost := perCluster.hostTable[hn]
		var newRecord bool
		if perHost == nil {
			perHost = new(hostInfo)
			perHost.oldestScannedTime = now
			perHost.oldestRecord = now
			perCluster.hostTable[hn] = perHost
			newRecord = true
		}

		if fromTime < perHost.oldestScannedTime {
			worklist = append(worklist, QueryArgs{
				QueryFilter: common.QueryFilter{
					HaveFrom: true,
					FromDate: time.Unix(fromTime-twoWeeks, 0).UTC(),
					HaveTo:   true,
					ToDate:   time.Unix(perHost.oldestScannedTime, 0).UTC(),
					Host:     []string{hn},
				},
			})
		}
		if !newRecord && toTime > perHost.youngestRecord && now-perHost.lastScan > oneHour {
			worklist = append(worklist, QueryArgs{
				QueryFilter: common.QueryFilter{
					HaveFrom: true,
					FromDate: time.Unix(perHost.youngestRecord, 0).UTC(),
					HaveTo:   true,
					ToDate:   nowt,
					Host:     []string{hn},
				},
			})
		}
	}

	return worklist
}

func (cdp *ConfigDataProvider) obtainAllFromCache(qa QueryArgs) (result []*NodeConfig) {
	result = make([]*NodeConfig, 0)

	perCluster := cdp.data
	perCluster.hostTableLock.Lock()
	defer perCluster.hostTableLock.Unlock()

	for _, hn := range qa.Host {
		perHost := perCluster.hostTable[hn]
		if perHost == nil {
			// Should we panic?
			continue
		}
		result = append(result,
			perHost.getLocked(qa.FromDate.UTC().Unix(), qa.ToDate.UTC().Unix())...)
	}

	return
}

func (cdp *ConfigDataProvider) obtainOneFromCache(host string, t int64) *NodeConfig {
	perCluster := cdp.data
	perCluster.hostTableLock.Lock()
	defer perCluster.hostTableLock.Unlock()

	perHost := perCluster.hostTable[host]
	if perHost == nil {
		// Should we panic?
		return nil
	}
	// I want the most recent <= t
	rs := perHost.records
	i := len(rs) - 1
	for i >= 0 && rs[i].Time > t {
		i--
	}
	if i >= 0 {
		return rs[i]
	}
	return nil
}

// Cache purging may be a ting.  What's the data load here?
//
// - suppose 3000 nodes across all the clusters
// - suppose node data updated once a day
// - suppose the max time period of real interest is on the order of days or at most a couple weeks
// - could look like one entry is as much as 128 bytes and pointerful (probably more)
// - => 14 * 3000 * 128 = 5.3MB (plus overhead for various spines)
//
// Cache purging could just set the entire array for a host to empty, or it could remove the oldest
// or youngest elements in some selection of tables.
//
// Probably here if the occupancy exceeds some limit, spin off a goroutine that takes the lock and
// removes some elements from the cache.  But it must be careful not to remove elements that may be
// used.  So probably there's a RWMutex and all consumer threads take this mutex on entry as readers
// and hold onto it while they do their thing.  Then the cache cleaner tries to take it as a writer
// and when it gets it it knows it can clean.
