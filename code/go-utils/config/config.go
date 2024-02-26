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

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"go-utils/hostglob"
)

type NodeMeta struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type NodeConfigRecord struct {
	// Full ISO timestamp of when the reading was taken (missing in older data)
	Timestamp string `json:"timestamp,omitempty"`

	// Arbitrary text, useful to somebody, may or may not be preserved by a processor (missing in
	// older data)
	Comment string `json:"comment,omitempty"`

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

type ClusterConfig struct {
	// Currently only one dimension
	nodes map[string]*NodeConfigRecord
}

func NewClusterConfig() *ClusterConfig {
	return &ClusterConfig{ nodes: make(map[string]*NodeConfigRecord) }
}

func (cc *ClusterConfig) Insert(r *NodeConfigRecord) {
	cc.nodes[r.Hostname] = r
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

	bytes, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &configInfo)
	if err != nil {
		return nil, fmt.Errorf("While unmarshaling config data: %w", err)
	}

	for _, c := range configInfo {
		if c.CpuCores == 0 || c.MemGB == 0 {
			return nil, fmt.Errorf("Nonsensical CPU/memory information for %s", c.Hostname)
		}
		if c.GpuCards == 0 && (c.GpuMemGB != 0 || c.GpuMemPct) {
			return nil, fmt.Errorf("Inconsistent GPU information for %s", c.Hostname)
		}
	}

	m, err := ExpandNodeConfigs(configInfo)
	if err != nil {
		return nil, err
	}
	return &ClusterConfig{ m }, nil
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
			panic("No way")
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
	outBytes, err := json.MarshalIndent(&records, "", " ")
	if err != nil {
		return err
	}
	// Ignore write errors here
	io.WriteString(os.Stdout, string(outBytes)+"\n")
	return nil
}
