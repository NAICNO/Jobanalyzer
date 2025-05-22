package db

import (
	"slices"
	"testing"
	"time"

	. "sonalyze/common"
)

func TestFilenames(t *testing.T) {
	theCfg, err := ReadConfigData(MakeConfigFilePath("testdata", "cluster1.uio.no"))
	if err != nil {
		t.Fatal(err)
	}
	if theCfg == nil {
		t.Fatal("nil config")
	}
	theDB, err := OpenPersistentCluster(MakeClusterDataPath("testdata", "cluster1.uio.no"), theCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer theDB.Close()

	// In these tests:
	// - from/to should not encompass the entire range of data in the database, there should be data
	//   for the selected hosts outside the range
	// - the globber should not match all hosts that have data in the selected range
	// - there should be proscribed files in the directories in that range
	// - there should be both new and old file name patterns
	// - note that proscribed files are only a matter if the globber is nil, so we also want to run
	//   with a nil globber
	// - there is special logic to round timestamps to date boundaries.  The `from` is rounded down
	//   to the start of that day.  The `to` is rounded up to the end of the day.  So we want to use
	//   odd timestamps for that reason

	// Ergo:
	// There's a directory at 04-11 with matching files
	// There's a directory at 05-04 with matching files
	// Some files are for n3.cluster1
	// There are proscribed files in various spots (bughunt.csv, cpuhog.csv)
	// There are both new and old file name schemes
	// There are files for all file types in all directories

	from, _ := time.Parse(time.RFC3339, "2025-04-12T07:16:00+02:00")
	to, _ := time.Parse(time.RFC3339, "2025-05-03T12:13:14+02:00")
	globber, _ := NewHosts(true, []string{"n[1-2].cluster1"})

	fs, err := theDB.SampleFilenames(from, to, nil)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(fs)
	expect := []string{
		"testdata/data/cluster1.uio.no/2025/04/12/0+sample-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/12/n1.cluster1.uio.no.csv",
		"testdata/data/cluster1.uio.no/2025/04/12/n3.cluster1.uio.no.csv",
		"testdata/data/cluster1.uio.no/2025/04/13/0+sample-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/13/0+sample-n3.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+sample-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+sample-n3.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/n1.cluster1.uio.no.csv",
		"testdata/data/cluster1.uio.no/2025/05/03/n1.cluster1.uio.no.csv",
	}
	if !slices.Equal(fs, expect) {
		t.Fatal(fs)
	}

	fs, err = theDB.SampleFilenames(from, to, globber)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(fs)
	expect = []string{
		"testdata/data/cluster1.uio.no/2025/04/12/0+sample-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/12/n1.cluster1.uio.no.csv",
		"testdata/data/cluster1.uio.no/2025/04/13/0+sample-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+sample-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/n1.cluster1.uio.no.csv",
		"testdata/data/cluster1.uio.no/2025/05/03/n1.cluster1.uio.no.csv",
	}
	if !slices.Equal(fs, expect) {
		t.Fatal(fs)
	}

	fs, err = theDB.SysinfoFilenames(from, to, nil)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(fs)
	expect = []string{
		"testdata/data/cluster1.uio.no/2025/04/12/0+sysinfo-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/12/sysinfo-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/12/sysinfo-n3.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/13/0+sysinfo-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/13/0+sysinfo-n3.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+sysinfo-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+sysinfo-n3.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/sysinfo-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/03/sysinfo-n1.cluster1.uio.no.json",
	}
	if !slices.Equal(fs, expect) {
		t.Fatal(fs)
	}

	fs, err = theDB.SysinfoFilenames(from, to, globber)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(fs)
	expect = []string{
		"testdata/data/cluster1.uio.no/2025/04/12/0+sysinfo-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/12/sysinfo-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/04/13/0+sysinfo-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+sysinfo-n2.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/02/sysinfo-n1.cluster1.uio.no.json",
		"testdata/data/cluster1.uio.no/2025/05/03/sysinfo-n1.cluster1.uio.no.json",
	}
	if !slices.Equal(fs, expect) {
		t.Fatal(fs)
	}

	fs, err = theDB.SacctFilenames(from, to)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(fs)
	expect = []string{
		"testdata/data/cluster1.uio.no/2025/04/12/0+job-slurm.json",
		"testdata/data/cluster1.uio.no/2025/04/12/slurm-sacct.csv",
		"testdata/data/cluster1.uio.no/2025/04/13/0+job-slurm.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+job-slurm.json",
		"testdata/data/cluster1.uio.no/2025/05/03/slurm-sacct.csv",
	}
	if !slices.Equal(fs, expect) {
		t.Fatal(fs)
	}

	fs, err = theDB.CluzterFilenames(from, to)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(fs)
	expect = []string{
		"testdata/data/cluster1.uio.no/2025/04/12/0+cluzter-slurm.json",
		"testdata/data/cluster1.uio.no/2025/04/13/0+cluzter-slurm.json",
		"testdata/data/cluster1.uio.no/2025/05/02/0+cluzter-slurm.json",
		"testdata/data/cluster1.uio.no/2025/05/03/0+cluzter-slurm.json",
	}
	if !slices.Equal(fs, expect) {
		t.Fatal(fs)
	}
}
