// This is logically a unit test for the printing of config information, but in actuality an
// integration test that tests the `config` verb since it uses the entire machinery underneath that
// verb.

package application

import (
	"testing"

	"sonalyze/cmd/configs"
)

func TestConfigs(t *testing.T) {
	testitConfigs(t, "timestamp,host,desc,xnode,cores,mem,gpus,gpumem,gpumempct")
	testitConfigs(
		t,
		"Timestamp,Hostname,Description,CrossNodeJobs,CpuCores,MemGB,GpuCards,GpuMemGB,GpuMemPct",
	)
}

func testitConfigs(t *testing.T, fields string) {
	var (
		configFilename = "testdata/config_format_test/test-config.json"
		expect         = []string{
			`,n1,"2x14 Intel Xeon Gold 5120 (hyperthreaded), 128GB, 4x NVIDIA RTX 2080 Ti @ 11GB",no,56,128,4,44,no`,
			`2024-10-31T12:00:00Z,n2,"1x14 Intel Xeon Gold 5120 (hyperthreaded), 256GB, 2x NVIDIA RTX 2080 Ti @ 11GB",yes,28,256,2,22,no`,
			``,
		}
	)
	var cc configs.ConfigCommand
	cc.ConfigFileArgs.ConfigFilename = configFilename
	cc.FormatArgs.Fmt = "csv,header," + fields
	testSimpleCommand(t, &cc, fields, expect)
}
