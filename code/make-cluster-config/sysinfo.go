package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path"

	"go-utils/config"
	"go-utils/filesys"
)

// Return a map from host to info containing the most recent record per host.
// We don't use the background information for anything here.

func readNodesFromSysinfo(_ map[string]*config.NodeConfigRecord) map[string]*config.NodeConfigRecord {
	rawInfo := readRawSysinfo()
	nodes := make(map[string]*config.NodeConfigRecord, 0)
	for _, infos := range rawInfo {
		// For the v2 format we can only have one timestamp, so take the latest always.
		var latest *config.NodeConfigRecord
		for _, info := range infos {
			if latest == nil || info.Timestamp > latest.Timestamp {
				latest = info
			}
		}
		if latest == nil {
			// Wow, weird
			continue
		}
		nodes[latest.Hostname] = latest
	}
	return nodes
}

// Return a map from host to an unordered list of records for the host.

func readRawSysinfo() map[string][]*config.NodeConfigRecord {
	files, err := filesys.EnumerateFiles(dataDir, from, to, "sysinfo-*.json")
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatalf("No sysinfo files found in %s", dataDir)
	}
	info := make(map[string][]*config.NodeConfigRecord)
	for _, fn := range files {
		input, err := os.Open(path.Join(dataDir, fn))
		if err != nil {
			log.Fatal(err)
		}
		// The sysinfo file has zero or more records in a row, but not wrapped in an array or
		// separated by anything more than space.  Thus, use a decoder to read the successive
		// records.
		dec := json.NewDecoder(input)
		for {
			var d config.NodeConfigRecord
			err := dec.Decode(&d)
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			info[d.Hostname] = append(info[d.Hostname], &d)
		}
	}
	return info
}
