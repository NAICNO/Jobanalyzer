// This is logically a unit test for the printing of node information, but in actuality an
// integration test that tests the `node` verb since it uses the entire machinery underneath that
// verb.

package application

import (
	"testing"

	"sonalyze/cmd/nodes"
)

func TestNode(t *testing.T) {
	testitNode(t, "timestamp,host,desc,cores,mem,gpus,gpumem,gpumempct")
	testitNode(t, "Timestamp,Hostname,Description,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct")
}

func testitNode(t *testing.T, fields string) {
	var (
		logFiles = []string{
			"testdata/node_format_test/sysinfo1.json",
			"testdata/node_format_test/sysinfo2.json"}
		expect = []string{
			`2024-11-01T00:00:01+01:00,ml1.hpc.uio.no,"2x14 (hyperthreaded) Intel(R) Xeon(R) Gold 5120 CPU @ 2.20GHz, 125 GiB, 3x NVIDIA GeForce RTX 2080 Ti @ 11GiB",56,125,3,33,no`,
			`2024-11-01T00:00:01+01:00,ml6.hpc.uio.no,"2x16 (hyperthreaded) AMD EPYC 7282 16-Core Processor, 251 GiB, 8x NVIDIA GeForce RTX 2080 Ti @ 11GiB",64,251,8,88,no`,
			``,
		}
	)
	var nc nodes.NodeCommand
	nc.DatabaseArgs.SetLogFiles(logFiles, "logfiles.cluster")
	nc.FormatArgs.Fmt = "csv,header," + fields
	testSimpleCommand(t, "node", &nc, fields, expect)
}
