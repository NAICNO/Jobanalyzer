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
	   The sonalyze remote server (required if not set in your ~/.sonalyze).  Normally
	   this will be an https: URL.

	-auth-file filename
	   The authorization file (required if not set in your ~/.sonalyze).  This is a
	   .netrc file or a file with username:password.

	-name cluster-name
	   The *canonical* cluster name to be used in the output (required)

	-desc description
	   The description of the cluster to be used in the output (required)

	-aliases alias,...
	   A list of short aliases for the cluster, to be used in the output

	-cross-node
	   Flag the cluster as being able to distribute jobs across multiple nodes.  This
	   is a bad hack but for now we have it.  Default false.  Normally, set it to true
	   on every cluster that runs Slurm.
*/
package main

// TODO: The code here is pretty much the same as that of ../make-cluster-config except for the use
// of the sonalyze server to provide sysinfo data.  We should merge the two programs.

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"go-utils/config"
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

const csvArgs = "host,desc,cores,mem,gpus,gpumem,timestamp"
const (
	hostIx = iota
	descIx
	coresIx
	memIx
	gpusIx
	gpuMemIx
	timestampIx
	numCsvFields
)

type node = config.NodeConfigRecord

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
	args := []string{
		"node",
		"-cluster", *cluster,
		"-fmt", "csv," + csvArgs,
		"-f", "4w",
		"-newest",
	}
	if *remote != "" {
		args = append(args, "-remote", *remote)
	}
	if *authFile != "" {
		args = append(args, "-auth-file", *authFile)
	}
	cmd := exec.Command(*sonalyze, args...)
	stderr := &strings.Builder{}
	cmd.Stderr = stderr
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(fmt.Sprintf("Sonalyze terminated with error: %v.  Command output:\n%s", err, stderr.String()))
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
		if len(rec) < numCsvFields {
			log.Fatal("Record too short: ", len(rec))
		}
		nodes = append(nodes, &node{
			Hostname:      rec[hostIx],
			Description:   rec[descIx],
			CrossNodeJobs: *crossNode,
			CpuCores:      int(number(rec[coresIx])),
			MemGB:         int(number(rec[memIx])),
			GpuCards:      int(number(rec[gpusIx])),
			GpuMemGB:      int(number(rec[gpuMemIx])),
			Timestamp:     rec[timestampIx],
		})
	}
	cfg := config.NewClusterConfig(
		2,
		*name,
		*description,
		parsedAliases,
		nil,
		nodes,
	)
	err = config.WriteConfigTo(os.Stdout, cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func number(s string) uint64 {
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return n
}
