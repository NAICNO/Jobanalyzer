package filedb

import (
	"slices"
	"strings"
	"testing"
	"time"

	. "sonalyze/common"
	"sonalyze/db/special"
)

var theDB *PersistentCluster

func getPersistentDB(t *testing.T, cluster string) *PersistentCluster {
	if theDB != nil {
		return theDB
	}
	var err error
	theCfg, err := special.ReadConfigData(special.MakeConfigFilePath("testdata", cluster))
	if err != nil {
		t.Fatal(err)
	}
	if theCfg == nil {
		t.Fatal("nil config")
	}
	theDB = NewPersistentCluster(special.MakeClusterDataPath("testdata", cluster), theCfg)
	return theDB
}

func TestFilenames(t *testing.T) {
	theDB := getPersistentDB(t, "cluster1.uio.no")

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

func TestData(t *testing.T) {
	theDB := getPersistentDB(t, "cluster1.uio.no")
	from, _ := time.Parse(time.RFC3339, "2025-04-12T07:16:00+02:00")
	to, _ := time.Parse(time.RFC3339, "2025-05-03T12:13:14+02:00")
	globber, _ := NewHosts(true, []string{"n[1-2].cluster1"})

	// There should be no gunk in the test data
	// We have six sample data files in the range, see test case for file names above

	sampleData, softErrors, err := theDB.ReadSamples(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(sampleData) != 6 {
		t.Fatal("Wrong number of data files", len(sampleData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var sampleNewFound, sampleOldFound bool
SampleLoop:
	for _, l := range sampleData {
		for _, d := range l {
			timestamp := time.Unix(int64(d.Timestamp), 0).UTC().Format(time.RFC3339)
			if strings.HasPrefix(timestamp, "2025-04-12") {
				sampleOldFound = true
			}
			if strings.HasPrefix(timestamp, "2025-04-13") {
				sampleNewFound = true
			}
			if sampleOldFound && sampleNewFound {
				break SampleLoop
			}
		}
	}
	if !sampleNewFound {
		t.Fatal("No sampleNew data read")
	}
	if !sampleOldFound {
		t.Fatal("No sampleOld data read")
	}

	loadData, softErrors, err := theDB.ReadLoadData(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(loadData) != 6 {
		t.Fatal("Wrong number of data files", len(loadData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var loadDataNewFound, loadDataOldFound bool
LoadDataLoop:
	for _, l := range loadData {
		for _, d := range l {
			timestamp := time.Unix(int64(d.Timestamp), 0).UTC().Format(time.RFC3339)
			if strings.HasPrefix(timestamp, "2025-04-12") {
				loadDataOldFound = true
			}
			if strings.HasPrefix(timestamp, "2025-04-13") {
				loadDataNewFound = true
			}
			if loadDataOldFound && loadDataNewFound {
				break LoadDataLoop
			}
		}
	}
	if !loadDataNewFound {
		t.Fatal("No loadDataNew data read")
	}
	if !loadDataOldFound {
		t.Fatal("No loadDataOld data read")
	}

	gpuData, softErrors, err := theDB.ReadGpuData(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(gpuData) != 6 {
		t.Fatal("Wrong number of data files", len(gpuData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var gpuDataNewFound, gpuDataOldFound bool
GpuDataLoop:
	for _, l := range gpuData {
		for _, d := range l {
			timestamp := time.Unix(int64(d.Timestamp), 0).UTC().Format(time.RFC3339)
			if strings.HasPrefix(timestamp, "2025-04-12") {
				gpuDataOldFound = true
			}
			if strings.HasPrefix(timestamp, "2025-04-13") {
				gpuDataNewFound = true
			}
			if gpuDataOldFound && gpuDataNewFound {
				break GpuDataLoop
			}
		}
	}
	if !gpuDataNewFound {
		t.Fatal("No gpuDataNew data read")
	}
	if !gpuDataOldFound {
		t.Fatal("No gpuDataOld data read")
	}

	sysinfoData, softErrors, err := theDB.ReadSysinfoData(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(sysinfoData) != 6 {
		t.Fatal("Wrong number of data files", len(sysinfoData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var sysinfoNewFound, sysinfoOldFound bool
SysinfoLoop:
	for _, l := range sysinfoData {
		for _, d := range l {
			if strings.HasPrefix(d.Timestamp, "2025-04-12") {
				sysinfoOldFound = true
			}
			if strings.HasPrefix(d.Timestamp, "2025-04-13") {
				sysinfoNewFound = true
			}
			if sysinfoOldFound && sysinfoNewFound {
				break SysinfoLoop
			}
		}
	}
	if !sysinfoNewFound {
		t.Fatal("No sysinfoNew data read")
	}
	if !sysinfoOldFound {
		t.Fatal("No sysinfoOld data read")
	}

	sacctData, softErrors, err := theDB.ReadSacctData(from, to, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(sacctData) != 5 {
		t.Fatal("Wrong number of data files", len(sacctData))
	}
	// Note, no timestamps in these data
	// At 2025-05-02 there is new-format data, user names start with "u"
	// at 2025-05-03 there is old-format data, user names start with "x"
	var sacctNewFound, sacctOldFound bool
SacctLoop:
	for _, l := range sacctData {
		for _, d := range l {
			if strings.HasPrefix(d.User.String(), "u") {
				sacctNewFound = true
			}
			if strings.HasPrefix(d.User.String(), "x") {
				sacctOldFound = true
			}
			if sacctOldFound && sacctNewFound {
				break SacctLoop
			}
		}
	}
	if !sacctNewFound {
		t.Fatal("No sacctNew data read")
	}
	if !sacctOldFound {
		t.Fatal("No sacctOld data read")
	}

	cluzterData, softErrors, err := theDB.ReadCluzterData(from, to, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(cluzterData) != 4 {
		t.Fatal("Wrong number of data files", len(cluzterData))
	}
	// At 2025-05-03 there is new-format data (the only kind)
	var cluzterFound bool
CluzterLoop:
	for _, l := range cluzterData {
		for _, d := range l {
			if strings.HasPrefix(string(d.Time), "2025-05-03T") {
				cluzterFound = true
				break CluzterLoop
			}
		}
	}
	if !cluzterFound {
		t.Fatal("No cluzter data read")
	}
}
