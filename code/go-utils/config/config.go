// Manage cluster configuration data.
//
// The cluster configuration is a two-dimensional map of JSON objects where each object represents
// the configuration of one or more nodes on the cluster at some point in time, with the fields of
// `NodeConfigRecord` below.
//
// Given a timestamp and a host, the config record can be found for the host closest to the
// timestamp.  Alternatively, the latest config record for a host can be found.
//
// At this time, the time dimension has not been implemented.
//
// See ../../../production/sonar-nodes/$CLUSTER/$CLUSTER-config.json for examples of config files.

// File formats.
//
// v2 file format:
//
// An object { ... } with the following named fields and value types:
//
//   name - string, the canonical name of the cluster
//   description - string, optional, arbitrary text describing the cluster
//   aliases - array of strings, optional, aliases / short names for the cluster
//   exclude-user - array of strings, optional, user names whose records should
//      be excluded when filtering records
//   nodes - array of objects, the list of nodes in the v1 format (see below)
//
// Any field name starting with '#' is reserved for arbitrary comments.
//
// The `exclude-user` option is a hack and is used to add post-hoc filtering of data (when Sonar
// should have filtered it to begin with, but didn't).  It is on purpose very limited, in contrast
// with e.g. a mechanism to add arbitrary arguments to the command line.  Additional filters, eg
// for command names, can be added as needed.
//
// v1 file format:
//
// An array [...] of objects { ... }, each with the following named fields and value types:
//
//   timestamp - string, optional, an RFC3339 timestamp for when the data were obtained
//   hostname - string, the fully qualified and unique host name of the node
//   description - string, optional, arbitrary text describing the node
//   cross_node_jobs - bool, optional, expressing that jobs on this node can be merged with
//                     jobs on other nodes in the same cluster where the flag is also set,
//                     because the job numbers come from the same cluster-wide source
//                     (typically slurm).  Also see the --batch option.
//   cpu_cores - integer, the number of hyperthreads
//   mem_gb - integer, the amount of main memory in gigabytes
//   gpu_cards - integer, the number of gpu cards on the node
//   gpumem_gb - integer, the amount of gpu memory in gigabytes across all cards
//   gpumem_pct - bool, optional, expressing a preference for the GPU memory reading
//
// Any field name starting with '#' is reserved for arbitrary comments.

package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"go-utils/hostglob"
	umaps "go-utils/maps"
)

type NodeMeta struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type NodeConfigRecord struct {
	// Full ISO timestamp of when the reading was taken (missing in older data)
	Timestamp string `json:"timestamp,omitempty"`

	// Name that host is known by on the cluster
	Hostname string `json:"hostname"`

	// End-user description, not parseable
	Description string `json:"description"`

	// True iff a job on this node can be merged with a job from a different node within an
	// appropriate time window if their job numbers are the same. (Missing in older data.)
	CrossNodeJobs bool `json:"cross_node_jobs,omitempty"`

	// Total number of cores x threads
	CpuCores int `json:"cpu_cores"`

	// GB of installed main RAM
	MemGB int `json:"mem_gb"`

	// Number of installed cards
	GpuCards int `json:"gpu_cards,omitempty"`

	// Total GPU memory across all cards
	GpuMemGB int `json:"gpumem_gb,omitempty"`

	// If true, use the percentage-of-memory-per-process figure from the card rather than a memory
	// measurement
	GpuMemPct bool `json:"gpumem_pct,omitempty"`

	// Carries additional information used by code generators.  This field is not intended to appear
	// in "production" configuration files, only in background files.
	Metadata []NodeMeta `json:"metadata,omitempty"`
}

type ClusterConfigV2Repr struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Aliases     []string            `json:"aliases,omitempty"`
	ExcludeUser []string            `json:"exclude-user,omitempty"`
	Nodes       []*NodeConfigRecord `json:"nodes"`
}

// Immutable
type ClusterConfig struct {
	Version     int
	Name        string
	Description string
	Aliases     []string
	ExcludeUser []string
	// Currently only one dimension of data
	nodes map[string]*NodeConfigRecord
}

func NewClusterConfig(
	version int,
	name, desc string,
	aliases, excludeUsers []string,
	nodes []*NodeConfigRecord,
) *ClusterConfig {
	nodemap := make(map[string]*NodeConfigRecord, len(nodes))
	for _, v := range nodes {
		nodemap[v.Hostname] = v
	}
	return &ClusterConfig{
		Version:     version,
		Name:        name,
		Description: desc,
		Aliases:     append([]string{}, aliases...),
		ExcludeUser: append([]string{}, excludeUsers...),
		nodes:       nodemap,
	}
}

// Finds the node configuration closest in time to the time stamp.  Returns nil if not found.
func (cc *ClusterConfig) LookupHostByTime(hostname string, timestamp string) *NodeConfigRecord {
	panic("No support for the time dimension of config yet")
}

// Returns the most recent node configuration for the node.  Returns nil if not found.
func (cc *ClusterConfig) LookupHost(hostname string) *NodeConfigRecord {
	if probe, found := cc.nodes[hostname]; found {
		return probe
	}
	return nil
}

// Return a fresh slice of all nodes in the config
func (cc *ClusterConfig) Hosts() []*NodeConfigRecord {
	return umaps.Values(cc.nodes)
}

// Returns the hosts that were defined within the time window.  With our current structure we don't
// have reliable time window information, so just return all hosts.
func (cc *ClusterConfig) HostsDefinedInTimeWindow(fromIncl, toExcl int64) []string {
	result := make([]string, 0)
	for _, n := range cc.nodes {
		result = append(result, n.Hostname)
	}
	return result
}

func (cc *ClusterConfig) HasCrossNodeJobs() bool {
	for _, n := range cc.nodes {
		if n.CrossNodeJobs {
			return true
		}
	}
	return false
}

// Get the system config if possible, returning a map from node names to node information.

func ReadConfig(configFilename string) (*ClusterConfig, error) {
	configFile, err := os.Open(configFilename)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	return ReadConfigFrom(configFile)
}

func ReadConfigFrom(input io.Reader) (*ClusterConfig, error) {
	var configInfo []*NodeConfigRecord
	bs, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(bytes.NewReader(bs))
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("While unmarshaling config data: %w", err)
	}
	config := new(ClusterConfig)
	if delim, ok := tok.(json.Delim); ok {
		switch delim {
		case '[':
			config.Version = 1
			config.Name = "v1 config"
			config.Description = "v1 config"
			err = json.Unmarshal(bs, &configInfo)
		case '{':
			var v2 ClusterConfigV2Repr
			err = json.Unmarshal(bs, &v2)
			config.Version = 2
			config.Name = v2.Name
			config.Description = v2.Description
			config.Aliases = v2.Aliases
			config.ExcludeUser = v2.ExcludeUser
			configInfo = v2.Nodes
		default:
			err = fmt.Errorf("Unexpected delimiter in JSON file %c", delim)
		}
	} else {
		err = fmt.Errorf("Unexpected non-delimiter in JSON file")
	}
	if err != nil {
		return nil, fmt.Errorf("While unmarshaling config data: %w", err)
	}

	for _, c := range configInfo {
		if c.CpuCores == 0 {
			return nil, fmt.Errorf("Zero or missing 'cpu_cores' in information for %s", c.Hostname)
		}
		if c.MemGB == 0 {
			return nil, fmt.Errorf("Zero or missing 'mem_gb' in information for %s", c.Hostname)
		}
		if c.GpuCards == 0 && (c.GpuMemGB != 0 || c.GpuMemPct) {
			return nil, fmt.Errorf("Inconsistent GPU information for %s", c.Hostname)
		}
	}

	m, err := ExpandNodeConfigs(configInfo)
	if err != nil {
		return nil, err
	}
	config.nodes = m
	return config, nil
}

func ExpandNodeConfigs(configInfo []*NodeConfigRecord) (map[string]*NodeConfigRecord, error) {
	moreInfo := []*NodeConfigRecord{}
	for _, c := range configInfo {
		expanded, err := hostglob.ExpandPattern(c.Hostname)
		if err != nil {
			return nil, err
		}
		switch len(expanded) {
		case 0:
			panic("Unexpected case")
		case 1:
			c.Hostname = expanded[0]
		default:
			c.Hostname = expanded[0]
			for i := 1; i < len(expanded); i++ {
				var d NodeConfigRecord = *c
				d.Hostname = expanded[i]
				moreInfo = append(moreInfo, &d)
			}
		}
	}
	configInfo = append(configInfo, moreInfo...)

	finalInfo := make(map[string]*NodeConfigRecord)
	for _, c := range configInfo {
		if _, found := finalInfo[c.Hostname]; found {
			return nil, fmt.Errorf("Duplicate host name in config: %s", c.Hostname)
		}
		finalInfo[c.Hostname] = c
	}

	return finalInfo, nil
}

// Sorting boilerplate.
type configSlice []*NodeConfigRecord

func (a configSlice) Len() int           { return len(a) }
func (a configSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a configSlice) Less(i, j int) bool { return a[i].Hostname < a[j].Hostname }

// Write the database.  The database is formatted text and we want it to be diffable, so the
// database is sorted on output.  At the moment we have only one record per host and the database is
// sorted in ascending hostname order.

func WriteConfigTo(output io.Writer, config *ClusterConfig) error {
	records := make([]*NodeConfigRecord, 0, len(config.nodes))
	for _, v := range config.nodes {
		records = append(records, v)
	}
	sort.Sort(configSlice(records))

	var err error
	var outBytes []byte
	if config.Version <= 1 {
		outBytes, err = json.MarshalIndent(&records, "", " ")
	} else {
		var v2repr ClusterConfigV2Repr
		v2repr.Name = config.Name
		v2repr.Description = config.Description
		v2repr.Aliases = config.Aliases
		v2repr.ExcludeUser = config.ExcludeUser
		v2repr.Nodes = records
		outBytes, err = json.MarshalIndent(&v2repr, "", " ")
	}
	if err != nil {
		return err
	}

	// Ignore write errors here
	io.WriteString(output, string(outBytes)+"\n")
	return nil
}
