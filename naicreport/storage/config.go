// Managing the "config" data.  These are going to become more complex over time,
// with optional and conditional parts.

package storage

import (
	"encoding/json"
	"io"
	"os"
)

type SystemConfig struct {
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
	CpuCores    int    `json:"cpu_cores"`
	MemGB       int    `json:"mem_gb"`
	GpuCards    int    `json:"gpu_cards"`
	GpuMemGB    int    `json:"gpumem_gb"`
	GpuMemPct   bool   `json:"gpumem_pct"`
}

// Get the system config if possible

func ReadConfig(configFilename string) ([]*SystemConfig, error) {
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
		return nil, err
	}

	return configInfo, nil
}
