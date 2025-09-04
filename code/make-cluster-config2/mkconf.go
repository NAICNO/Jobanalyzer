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
	parsedAliases := []string{}
	if *aliases != "" {
		parsedAliases = strings.Split(*aliases, ",")
	}

	// run sonalyze with the args + -fmt csv,host,desc,cores,mem,gpus,gpumem -newest

	args := []string{"node", "-cluster", *cluster, "-fmt", "csv,host,desc,cores,mem,gpus,gpumem", "-newest"}
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
