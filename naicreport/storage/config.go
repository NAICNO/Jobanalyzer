// Managing the "config" data.  These are going to become more complex over time,
// with optional and conditional parts.

package storage

import (
	"encoding/json"
	"errors"
	"fmt"
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

	for _, c := range configInfo {
		if c.CpuCores == 0 || c.MemGB == 0 {
			return nil, errors.New(fmt.Sprintf("Nonsensical CPU/memory information for %s", c.Hostname))
		}
		if c.GpuCards == 0 && (c.GpuMemGB != 0 || c.GpuMemPct) {
			return nil, errors.New(fmt.Sprintf("Inconsistent GPU information for %s", c.Hostname))
		}
	}

	moreInfo := []*SystemConfig{}
	for _, c := range configInfo {
		expanded := ExpandPatterns(c.Hostname)
		switch len(expanded) {
		case 0:
			panic("No way")
		case 1:
			c.Hostname = expanded[0]
		default:
			c.Hostname = expanded[0]
			for i := 1; i < len(expanded); i++ {
				moreInfo = append(moreInfo, &SystemConfig{
					Hostname:    expanded[i],
					Description: c.Description,
					CpuCores:    c.CpuCores,
					MemGB:       c.MemGB,
					GpuCards:    c.GpuCards,
					GpuMemGB:    c.GpuMemGB,
					GpuMemPct:   c.GpuMemPct,
				})
			}
		}
	}
	configInfo = append(configInfo, moreInfo...)

	names := make(map[string]bool)
	for _, c := range configInfo {
		if _, found := names[c.Hostname]; found {
			return nil, errors.New(fmt.Sprintf("Duplicate host name in config: %s", c.Hostname))
		}
		names[c.Hostname] = true
	}

	return configInfo, nil
}
