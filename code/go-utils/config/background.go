// "Background" information is used when we generate cluster configurations from other data: it
// provides any missing data.
//
// A background file is currently a JSON array datum each with a (partial) NodeConfigRecord, where
// each host name may be a host pattern.  The information is expanded into a map from a hostname to
// the (partial) background information for the host.
//
// Some background information is "deep background" and does not represent eg hosts that were not
// found by sysinfo, and their information should not be emitted into cluster config files.  These
// records are marked with cpu_cores==0 or mem_gb==0.

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadBackgroundFile(filename string) (map[string]*NodeConfigRecord, error) {
	input, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Opening background data: %w", err)
	}
	defer input.Close()
	return ReadBackground(input)
}

func ReadBackground(input io.Reader) (map[string]*NodeConfigRecord, error) {
	bgBytes, err := io.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("Reading background data: %w", err)
	}
	var bgArray []*NodeConfigRecord
	err = json.Unmarshal(bgBytes, &bgArray)
	if err != nil {
		return nil, fmt.Errorf("Unmarshaling background data: %w", err)
	}
	return ExpandNodeConfigs(bgArray)
}
