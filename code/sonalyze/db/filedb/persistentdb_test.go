package filedb

import (
	"slices"
	"strings"
	"testing"
	"time"

	. "sonalyze/common"
	"sonalyze/db/repr"
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

// TODO: To make these tests stronger we could pre-declare variables receiving the data (of many
// kinds), to check that types are right.

func TestData(t *testing.T) {
	theDB := getPersistentDB(t, "cluster1.uio.no")
	from, _ := time.Parse(time.RFC3339, "2025-04-12T07:16:00+02:00")
	to, _ := time.Parse(time.RFC3339, "2025-05-03T12:13:14+02:00")
	globber, _ := NewHosts(true, []string{"n[1-2].cluster1"})

	var softErrors int
	var err error

	// There should be no gunk in the test data
	// We have six sample data files in the range, see test case for file names above

	var sampleData [][]*repr.Sample
	sampleData, softErrors, err = theDB.ReadProcessSamples(from, to, globber, true)
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

	var cpuSampleData [][]*repr.CpuSamples
	cpuSampleData, softErrors, err = theDB.ReadCpuSamples(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(cpuSampleData) != 6 {
		t.Fatal("Wrong number of data files", len(cpuSampleData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var cpuSampleDataNewFound, cpuSampleDataOldFound bool
CpuSampleDataLoop:
	for _, l := range cpuSampleData {
		for _, d := range l {
			timestamp := time.Unix(int64(d.Timestamp), 0).UTC().Format(time.RFC3339)
			if strings.HasPrefix(timestamp, "2025-04-12") {
				cpuSampleDataOldFound = true
			}
			if strings.HasPrefix(timestamp, "2025-04-13") {
				cpuSampleDataNewFound = true
			}
			if cpuSampleDataOldFound && cpuSampleDataNewFound {
				break CpuSampleDataLoop
			}
		}
	}
	if !cpuSampleDataNewFound {
		t.Fatal("No cpuSampleDataNew data read")
	}
	if !cpuSampleDataOldFound {
		t.Fatal("No cpuSampleDataOld data read")
	}

	var gpuSampleData [][]*repr.GpuSamples
	gpuSampleData, softErrors, err = theDB.ReadGpuSamples(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(gpuSampleData) != 6 {
		t.Fatal("Wrong number of data files", len(gpuSampleData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var gpuSampleDataNewFound, gpuSampleDataOldFound bool
GpuSampleDataLoop:
	for _, l := range gpuSampleData {
		for _, d := range l {
			timestamp := time.Unix(int64(d.Timestamp), 0).UTC().Format(time.RFC3339)
			if strings.HasPrefix(timestamp, "2025-04-12") {
				gpuSampleDataOldFound = true
			}
			if strings.HasPrefix(timestamp, "2025-04-13") {
				gpuSampleDataNewFound = true
			}
			if gpuSampleDataOldFound && gpuSampleDataNewFound {
				break GpuSampleDataLoop
			}
		}
	}
	if !gpuSampleDataNewFound {
		t.Fatal("No gpuSampleDataNew data read")
	}
	if !gpuSampleDataOldFound {
		t.Fatal("No gpuSampleDataOld data read")
	}

	var nodeData [][]*repr.SysinfoNodeData
	nodeData, softErrors, err = theDB.ReadSysinfoNodeData(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(nodeData) != 6 {
		t.Fatal("Wrong number of data files", len(nodeData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var nodeNewFound, nodeOldFound bool
NodeLoop:
	for _, l := range nodeData {
		for _, d := range l {
			if strings.HasPrefix(d.Time, "2025-04-12") {
				nodeOldFound = true
			}
			if strings.HasPrefix(d.Time, "2025-04-13") {
				nodeNewFound = true
			}
			if nodeOldFound && nodeNewFound {
				break NodeLoop
			}
		}
	}
	if !nodeNewFound {
		t.Fatal("No new node data read")
	}
	if !nodeOldFound {
		t.Fatal("No old node data read")
	}

	var cardData [][]*repr.SysinfoCardData
	cardData, softErrors, err = theDB.ReadSysinfoCardData(from, to, globber, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(cardData) != 6 {
		t.Fatal("Wrong number of data files", len(cardData))
	}
	// at 2025-04-12 there is old-format data for n1.cluster1
	// at 2025-04-13 there is new-format data for n1.cluster1
	var cardNewFound, cardOldFound bool
CardLoop:
	for _, l := range cardData {
		for _, d := range l {
			if strings.HasPrefix(d.Time, "2025-04-12") {
				cardOldFound = true
			}
			if strings.HasPrefix(d.Time, "2025-04-13") {
				cardNewFound = true
			}
			if cardNewFound && cardOldFound {
				break CardLoop
			}
		}
	}
	if !cardNewFound {
		t.Fatal("No new card data read")
	}
	if !cardOldFound {
		t.Fatal("No old card data read")
	}

	var sacctData [][]*repr.SacctInfo
	sacctData, softErrors, err = theDB.ReadSacctData(from, to, true)
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

	var slurmAttrData [][]*repr.CluzterAttributes
	slurmAttrData, softErrors, err = theDB.ReadCluzterAttributeData(from, to, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(slurmAttrData) != 4 {
		t.Fatal("Wrong number of data files", len(slurmAttrData))
	}
	// At 2025-05-03 there is new-format data (the only kind)
	var attrFound bool
AttrLoop:
	for _, l := range slurmAttrData {
		for _, d := range l {
			if strings.HasPrefix(string(d.Time), "2025-05-03T") {
				attrFound = true
				break AttrLoop
			}
		}
	}
	if !attrFound {
		t.Fatal("No cluzter attribute data read")
	}

	var slurmNodeData [][]*repr.CluzterNodes
	slurmNodeData, softErrors, err = theDB.ReadCluzterNodeData(from, to, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(slurmNodeData) != 4 {
		t.Fatal("Wrong number of data files", len(slurmAttrData))
	}
	// At 2025-05-03 there is new-format data (the only kind)
	var nodeFound bool
SlurmNodeLoop:
	for _, l := range slurmNodeData {
		for _, d := range l {
			if strings.HasPrefix(string(d.Time), "2025-05-03T") {
				nodeFound = true
				break SlurmNodeLoop
			}
		}
	}
	if !nodeFound {
		t.Fatal("No cluzter node data read")
	}

	var slurmPartitionData [][]*repr.CluzterPartitions
	slurmPartitionData, softErrors, err = theDB.ReadCluzterPartitionData(from, to, true)
	if err != nil {
		t.Fatal(err)
	}
	if softErrors != 0 {
		t.Fatal("Soft errors", softErrors)
	}
	if len(slurmPartitionData) != 4 {
		t.Fatal("Wrong number of data files", len(slurmAttrData))
	}
	// At 2025-05-03 there is new-format data (the only kind)
	var partitionFound bool
SlurmPartitionLoop:
	for _, l := range slurmPartitionData {
		for _, d := range l {
			if strings.HasPrefix(string(d.Time), "2025-05-03T") {
				partitionFound = true
				break SlurmPartitionLoop
			}
		}
	}
	if !partitionFound {
		t.Fatal("No cluzter partition data read")
	}
}
