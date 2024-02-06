// Manage cluster configuration data.
//
// The cluster configuration is an array of JSON objects where each object represents the
// configuration of one or more nodes on the cluster, with the fields of `SystemConfig` below.
//
// The configuration is currently time-invariant.
//
// (These data are going to become more complex over time, with optional and conditional parts.)

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"go-utils/hostglob"
)

type SystemMeta struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type SystemConfig struct {
	// Name that host is known by on the cluster
	Hostname string `json:"hostname"`

	// End-user description, not parseable
	Description string `json:"description"`

	// True iff a job on this node can be merged with a job from a different node within an
	// appropriate time window if their job numbers are the same.
	CrossNodeJobs bool `json:"cross_node_jobs,omitempty"`

	// Total number of cores x threads
	CpuCores int `json:"cpu_cores"`

	// GB of installed main RAM
	MemGB int `json:"mem_gb"`

	// Number of installed cards
	GpuCards int `json:"gpu_cards,omitempty"`

	// Total GPU memory across all cards
	GpuMemGB int `json:"gpumem_gb,omitempty"`

	// If true, use the percentage-of-memory-per-process figure from the card
	// rather than a memory measurement
	GpuMemPct bool `json:"gpumem_pct,omitempty"`

	// Carries additional information used by code generators
	Metadata []SystemMeta `json:"metadata,omitempty"`
}

// Get the system config if possible, returning a map from expanded host name to information for the
// host.  If `lenient` is true then the checks for CpuCores and MemGB are omitted.

func ReadConfig(configFilename string, lenient bool) (map[string]*SystemConfig, error) {
	var configInfo []*SystemConfig

	configFile, err := os.Open(configFilename)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &configInfo)
	if err != nil {
		return nil, fmt.Errorf("While unmarshaling config data: %w", err)
	}

	for _, c := range configInfo {
		if !lenient {
			if c.CpuCores == 0 || c.MemGB == 0 {
				return nil, fmt.Errorf("Nonsensical CPU/memory information for %s", c.Hostname)
			}
		}
		if c.GpuCards == 0 && (c.GpuMemGB != 0 || c.GpuMemPct) {
			return nil, fmt.Errorf("Inconsistent GPU information for %s", c.Hostname)
		}
	}

	moreInfo := []*SystemConfig{}
	for _, c := range configInfo {
		expanded := hostglob.ExpandPatterns(c.Hostname)
		switch len(expanded) {
		case 0:
			panic("No way")
		case 1:
			c.Hostname = expanded[0]
		default:
			c.Hostname = expanded[0]
			for i := 1; i < len(expanded); i++ {
				var d SystemConfig = *c
				d.Hostname = expanded[i]
				moreInfo = append(moreInfo, &d)
			}
		}
	}
	configInfo = append(configInfo, moreInfo...)

	finalInfo := make(map[string]*SystemConfig)
	for _, c := range configInfo {
		if _, found := finalInfo[c.Hostname]; found {
			return nil, fmt.Errorf("Duplicate host name in config: %s", c.Hostname)
		}
		finalInfo[c.Hostname] = c
	}

	return finalInfo, nil
}
