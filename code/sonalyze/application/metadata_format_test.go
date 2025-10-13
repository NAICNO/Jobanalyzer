// This is logically a unit test for the printing of metadata information, but in actuality an
// integration test that tests the `metadata` verb since it uses the entire machinery underneath that
// verb.

package application

import (
	"testing"

	"sonalyze/cmd/metadata"
)

func TestMetadata(t *testing.T) {
	testitMetadata(t, "host,earliest,latest")
	testitMetadata(t, "Hostname,Earliest,Latest")
}

func testitMetadata(t *testing.T, fields string) {
	var (
		expect = []string{
			`ml6.hpc.uio.no,2024-10-31 00:00,2024-10-31 00:30`,
			``,
		}
	)
	var mc metadata.MetadataCommand
	mc.DatabaseArgs.SetLogFiles([]string{"testdata/metadata_format/test.csv"}, "logfiles.cluster")
	mc.FormatArgs.Fmt = "csv,header," + fields
	mc.Bounds = true
	testSampleAnalysisCommand(t, &mc, fields, expect)
}
