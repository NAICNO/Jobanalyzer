/*
Mkconf creates a Jobanalyzer config file from sysinfo data.

It runs sonalyze with the given arguments to collect sysinfo data, and produces a
JSON configuration object for the cluster on standard output.

Usage:

	mkconf flag ...

The flags are:

	-sonalyze sonalyze-path
	   The path to the sonalyze executable (required)

	-cluster cluster-name
	   The name (could be short name) of the cluster to extract sysinfo for (required)

	-remote hostname
	   The sonalyze remote server (required if not set in your ~/.sonalyze)

	-auth-file filename
	   The authorization .netrc file (required if not set in your ~/.sonalyze)

	-name cluster-name
	   The *canonical* cluster name to be used in the output (required)

	-desc description
	   The description of the cluster to be used in the outpu (required)

	-aliases alias,...
	   A list of short aliases for the cluster (optional)

	-cross-node
	   Flag the cluster as being able to distribute jobs across multiple nodes.  This
	   is a bad hack but for now we have it.  Default false.
*/
package main

import (
	"bytes"
	"cmp"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// TODO: We want a -coalesce flag to coalesce nodes with the same descriptions and compatible
// (compressible) names, but it's not super important.  It saves a little memory in the backend on
// the really big systems, and it makes the config file smaller (the config file for Betzy is over
// 300KB).

var (
	sonalyze    = flag.String("sonalyze", "", "Path to sonalyze `executable`")
	cluster     = flag.String("cluster", "", "The `cluster name` for which we're generating")
	authFile    = flag.String("auth-file", "", "Path to `netrc-file` authorizing the query to the remote")
	remote      = flag.String("remote", "", "`hostname` of sonalyze server")
	name        = flag.String("name", "", "Canonical `cluster-name`")
	description = flag.String("desc", "", "Cluster `description`")
	aliases     = flag.String("aliases", "", "Comma-separated list of `aliases`")
	crossNode   = flag.Bool("cross-node", false, "Cross-node jobs are allowed")
)

// CSV fields
const (
	hostIx = iota
	descIx
	coresIx
	memIx
	gpusIx
	gpuMemIx
	timestampIx
	lastField
)

// JSON structure
type envelope struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases"`
	Description string   `json:"description"`
	Nodes       []*node  `json:"nodes"`
}

type node struct {
	Hostname    string `json:"hostname"`
	Description string `json:"description"`
	CrossNode   bool   `json:"cross_node_jobs,omitempty"`
	CpuCores    uint64 `json:"cpu_cores"`
	MemGB       uint64 `json:"mem_gb"`
	GpuCards    uint64 `json:"gpu_cards,omitempty"`
	GpuMemGB    uint64 `json:"gpumem_gb,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

func main() {
	flag.Parse()
	if *sonalyze == "" {
		log.Fatal("Missing -sonalyze")
	}
	if *cluster == "" {
		log.Fatal("Missing -cluster")
	}
	if *name == "" {
		log.Fatal("Missing -name")
	}
	if *description == "" {
		log.Fatal("Missing -desc")
	}
	var parsedAliases []string
	if *aliases != "" {
		parsedAliases = strings.Split(*aliases, ",")
		for i, v := range parsedAliases {
			parsedAliases[i] = strings.TrimSpace(v)
		}
	}
	args := []string{"node", "-cluster", *cluster, "-fmt", "csv,host,desc,cores,mem,gpus,gpumem,timestamp", "-f", "4w", "-newest"}
	if *remote != "" {
		args = append(args, "-remote", *remote)
	}
	if *authFile != "" {
		args = append(args, "-authfile", *authFile)
	}
	out, err := exec.Command(*sonalyze, args...).Output()
	if err != nil {
		log.Fatal(err)
	}
	rdr := csv.NewReader(bytes.NewReader(out))
	nodes := make([]*node, 0)
	for {
		rec, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if len(rec) < lastField {
			log.Fatal("Record too short")
		}
		nodes = append(nodes, &node{
			Hostname:    rec[hostIx],
			Description: rec[descIx],
			CrossNode:   *crossNode,
			CpuCores:    number(rec[coresIx]),
			MemGB:       number(rec[memIx]),
			GpuCards:    number(rec[gpusIx]),
			GpuMemGB:    number(rec[gpuMemIx]),
			Timestamp:   rec[timestampIx],
		})
	}
	slices.SortFunc(nodes, func(a, b *node) int {
		return cmp.Compare(a.Hostname, b.Hostname)
	})
	out, err = json.MarshalIndent(
		&envelope{
			Name:        *name,
			Aliases:     parsedAliases,
			Description: *description,
			Nodes:       nodes,
		}, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(out)
}

func number(s string) uint64 {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return n
}
