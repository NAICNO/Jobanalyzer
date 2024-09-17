// Unit test clusterstore, persistentcluster, transientcluster, logfile logic
//
// This tests only single-threaded accesses to the store.

package db

import (
	"io/fs"
	"os"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"go-utils/filesys"
	"go-utils/hostglob"
	uslices "go-utils/slices"
	"go-utils/status"
	. "sonalyze/common"
)

const (
	verbose = true
)

var (
	// MT: Constant after initialization; immutable
	theClusterDir = "../../tests/sonarlog/whitebox-tree"
	theFiles      = []string{
		"../../tests/sonarlog/parse-data1.csv",
		"../../tests/sonarlog/parse-data2.csv",
	}
)

func tmpCopyTree(srcDir string) string {
	targetDir, err := os.MkdirTemp("", "clusterstore")
	if err != nil {
		panic("Can't create tempdir")
	}
	filesys.CopyDir(srcDir, targetDir)
	return targetDir
}

func TestOpenClose(t *testing.T) {
	pc, err := OpenPersistentCluster(theClusterDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	qc, err := OpenPersistentCluster(theClusterDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if pc != qc {
		t.Fatal("Object identity")
	}

	// Closing should prevent more dirs from being opened
	Close()
	_, err = OpenPersistentCluster(theClusterDir, nil)
	if err != ClusterClosedErr {
		t.Fatal("Should be closed")
	}

	// This should work even after Close() because file clusters are independent of the cluster
	// store.
	fs, err := OpenTransientSampleCluster(theFiles, nil)
	if err != nil {
		t.Fatal(err)
	}
	gs, err := OpenTransientSampleCluster(theFiles, nil)
	if err != nil {
		t.Fatal(err)
	}
	if fs == gs {
		t.Fatal("Object identity")
	}

	unsafeResetClusterStore()
}

func TestTransientSampleFilenames(t *testing.T) {
	fs, err := OpenTransientSampleCluster(theFiles, nil)
	if err != nil {
		t.Fatal(err)
	}
	var d time.Time
	h, _ := hostglob.NewGlobber(false, []string{"a"})
	// The parameters should be ignored here and the names returned should
	// be exactly the input names.
	names, _ := fs.SampleFilenames(d, d, h)
	if !reflect.DeepEqual(names, theFiles) {
		t.Fatal(names, theFiles)
	}
}

func TestTransientSampleRead(t *testing.T) {
	fs, err := OpenTransientSampleCluster(theFiles, nil)
	if err != nil {
		t.Fatal(err)
	}
	// The parameters are ignored here
	var d time.Time
	sampleBlobs, dropped, err := fs.ReadSamples(d, d, nil, verbose)
	if err != nil {
		t.Fatal(err)
	}
	_ = dropped
	n := 0
	for _, s := range sampleBlobs {
		n += len(s)
	}
	if n != 7 {
		t.Fatal("Length", sampleBlobs)
	}
}

func TestPersistentSampleFilenames(t *testing.T) {
	pc, err := OpenPersistentCluster(theClusterDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	names, err := pc.SampleFilenames(
		time.Date(2023, 05, 28, 12, 37, 55, 0, time.UTC),
		time.Date(2023, 05, 31, 23, 0, 12, 0, time.UTC),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(sort.StringSlice(names))
	d := path.Clean(theClusterDir)
	// Exclude 05/29 because it's not a directory
	// Exclude 05/31/c.txt because it's not csv
	// Exclude 05/31/bughunt.csv because it's proscribed
	expect := []string{
		path.Join(d, "2023/05/28/a.csv"),
		path.Join(d, "2023/05/28/b.csv"),
		path.Join(d, "2023/05/30/a.csv"),
		path.Join(d, "2023/05/30/b.csv"),
		path.Join(d, "2023/05/31/a.csv"),
		path.Join(d, "2023/05/31/b.csv"),
	}
	if !reflect.DeepEqual(names, expect) {
		t.Fatal(names, expect)
	}

	h, _ := hostglob.NewGlobber(true, []string{"a"})
	names, err = pc.SampleFilenames(
		time.Date(2023, 05, 28, 12, 37, 55, 0, time.UTC),
		time.Date(2023, 05, 31, 23, 0, 12, 0, time.UTC),
		h,
	)
	if err != nil {
		t.Fatal(err)
	}
	sort.Sort(sort.StringSlice(names))
	expect = []string{
		path.Join(d, "2023/05/28/a.csv"),
		path.Join(d, "2023/05/30/a.csv"),
		path.Join(d, "2023/05/31/a.csv"),
	}
	if !reflect.DeepEqual(names, expect) {
		t.Fatal(names, expect)
	}
}

func TestPersistentSampleRead(t *testing.T) {
	pc, err := OpenPersistentCluster(theClusterDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	sampleBlobs, dropped, err := pc.ReadSamples(
		time.Date(2023, 05, 28, 12, 37, 55, 0, time.UTC),
		time.Date(2023, 05, 31, 23, 0, 12, 0, time.UTC),
		nil,
		verbose,
	)
	if err != nil {
		t.Fatal(err)
	}

	// 05/28/a.csv has a field named "yabbadabbadoo" and the field is dropped
	if dropped != 1 {
		t.Fatal("Dropped", dropped)
	}
	n := 0
	for _, s := range sampleBlobs {
		n += len(s)
	}
	if n != 7 {
		t.Fatal("Length", sampleBlobs)
	}
}

func TestPersistentSysinfoRead(t *testing.T) {
	pc, err := OpenPersistentCluster(theClusterDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	// 5/28 "a" should have one record
	// 5/28 "b" should not exist
	// 5/30 "a" should exist but be empty
	// 5/30 "b" should have one record
	// 5/31 "a" should have two records, not equal
	// 5/32 "b" should have two records, equal
	recordBlobs, dropped, err := pc.ReadSysinfo(
		time.Date(2023, 05, 28, 12, 37, 55, 0, time.UTC),
		time.Date(2023, 05, 31, 23, 0, 12, 0, time.UTC),
		nil,
		verbose,
	)
	if err != nil {
		t.Fatal(err)
	}
	if dropped != 0 {
		t.Fatal("Dropped", dropped)
	}
	n := 0
	for _, rs := range recordBlobs {
		n += len(rs)
	}
	if n != 6 {
		t.Fatal("Length", recordBlobs)
	}
}

func TestPersistentSampleAppend(t *testing.T) {
	d := tmpCopyTree(theClusterDir)
	defer os.RemoveAll(d)

	pc, err := OpenPersistentCluster(d, nil)
	if err != nil {
		t.Fatal(err)
	}

	// I think what we could do here is do multiple adds to multiple files and then close and then
	// check that everything looks good.
	//
	// We might try to interleave appending to various files, too - just to have done it.

	l1 := "v=0.11.0,time=2023-05-28T14:30:00+02:00,host=a,cores=6,user=larstha,job=249151,pid=11090,cmd=larceny,cpu%=100,cpukib=113989888"
	l2 := "v=0.11.0,time=2023-05-28T14:35:00+02:00,host=a,cores=8,user=lth,job=49151,pid=111090,cmd=flimflam,cpu%=100,cpukib=113989888"
	pc.AppendSamplesAsync(
		"a",
		"2023-05-28T14:30:00+02:00",
		l1+"\n",
	)
	pc.AppendSamplesAsync(
		"a",
		"2023-05-28T14:35:00+02:00",
		l2,
	)

	pc.Close()

	lines, _ := filesys.FileLines(path.Join(d, "2023/05/28/a.csv"))
	if lines[len(lines)-2] != l1 || lines[len(lines)-1] != l2 {
		x := lines[len(lines)-2]
		y := lines[len(lines)-1]
		t.Fatalf("Lines don't match\n<%s>\n<%s>", x, y)
	}
}

func TestPersistentSysinfoAppend(t *testing.T) {
	d := tmpCopyTree(theClusterDir)
	defer os.RemoveAll(d)

	pc, err := OpenPersistentCluster(d, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Existing nonempty file
	pc.AppendSysinfoAsync(
		"a",
		"2023-05-28T16:00:01+02:00",
		`{
  "timestamp": "2023-05-28T16:00:01+02:00",
  "hostname": "a",
  "description": "18x14 (hyperthreaded) Intel(R) Xeon(R) Gold 5120 CPU @ 2.20GHz, 125 GiB, 5x NVIDIA GeForce RTX 2080 Ti @ 11GiB",
  "cpu_cores": 252,
  "mem_gb": 125,
  "gpu_cards": 5,
  "gpumem_gb": 55
}`,
	)

	// New file in existing directory
	pc.AppendSysinfoAsync(
		"c",
		"2023-05-28T16:00:01+02:00",
		`{
  "timestamp": "2023-05-28T16:00:01+02:00",
  "hostname": "c",
  "description": "18x14 (hyperthreaded) Intel(R) Xeon(R) Gold 5120 CPU @ 2.20GHz, 125 GiB, 4x NVIDIA GeForce RTX 2080 Ti @ 11GiB",
  "cpu_cores": 252,
  "mem_gb": 125,
  "gpu_cards": 4,
  "gpumem_gb": 44
}`,
	)

	// New file in new directory
	pc.AppendSysinfoAsync(
		"d",
		"2024-04-12T16:00:01+02:00",
		`{
  "timestamp": "2024-04-12T16:00:01+02:00",
  "hostname": "d",
  "description": "18x14 (hyperthreaded) Intel(R) Xeon(R) Gold 5120 CPU @ 2.20GHz, 125 GiB, 4x NVIDIA GeForce RTX 2080 Ti @ 11GiB",
  "cpu_cores": 252,
  "mem_gb": 125,
  "gpu_cards": 4,
  "gpumem_gb": 44
}`,
	)

	// This also tests that reading without flushing sees the new data.  Technically we're allowed
	// to not see the data - a synchronous flush is technically required - and if we ever implement
	// that path then this test will need to have a FlushSync() call before the read.

	recordBlobs, _, err := pc.ReadSysinfo(
		time.Date(2023, 05, 28, 12, 37, 55, 0, time.UTC),
		time.Date(2023, 05, 28, 23, 0, 12, 0, time.UTC),
		nil,
		verbose,
	)
	if err != nil {
		t.Fatal(err)
	}
	// In the original data we had nonempty sysinfo only for "a" on "2023/05/28", and only one
	// record, so now we should have 3 in the first window and 1 in the second window.
	n := 0
	for _, rs := range recordBlobs {
		n += len(rs)
	}
	if n != 3 {
		t.Fatal("Length", recordBlobs)
	}

	recordBlobs2, _, err := pc.ReadSysinfo(
		time.Date(2024, 01, 01, 12, 37, 55, 0, time.UTC),
		time.Date(2024, 05, 01, 23, 0, 12, 0, time.UTC),
		nil,
		verbose,
	)
	if err != nil {
		t.Fatal(err)
	}
	n = 0
	for _, rs := range recordBlobs2 {
		n += len(rs)
	}
	if n != 1 {
		t.Fatal("Length", recordBlobs)
	}

	pc.Close()

	// Check that new files exist
	fs := os.DirFS(pc.dataDir).(fs.StatFS)
	_, err = fs.Stat("2023/05/28/sysinfo-c.json")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fs.Stat("2024/04/12/sysinfo-d.json")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPersistentSampleFlush(t *testing.T) {
	d := tmpCopyTree(theClusterDir)
	defer os.RemoveAll(d)

	pc, err := OpenPersistentCluster(d, nil)
	if err != nil {
		t.Fatal(err)
	}

	// No changes pending, so this should do nothing
	pc.FlushAsync()

	l1 := "v=0.11.0,time=2024-02-13T14:30:00+02:00,host=a,cores=6,user=larstha,job=249151,pid=11090,cmd=larceny,cpu%=100,cpukib=113989888"
	pc.AppendSamplesAsync(
		"c",
		"2024-02-13T14:30:00+02:00",
		l1+"\n",
	)

	// At the moment an async flush is sync, so we should see the effect immediately.  But sleep 1s
	// just to make it interesting.
	pc.FlushAsync()
	time.Sleep(1 * time.Second)

	lines, err := filesys.FileLines(path.Join(d, "2024/02/13/c.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if lines[len(lines)-1] != l1 {
		t.Fatalf("Lines don't match\n<%s>", lines[len(lines)-1])
	}
}

// Caching tests
//
// We can listen for Info messages about cache behavior:
//
//   "Caching <filename> size <x>"
//   "Purging <filename> b/c <reason>
//   "Budget <current-budget>"
//   "Cache hit <filename>"

type CacheListener struct /* implements status.UnderlyingLogger */ {
	sync.Mutex
	msgs []string
}

func (cl *CacheListener) Info(m string) error {
	cl.Lock()
	defer cl.Unlock()
	cl.msgs = append(cl.msgs, m)
	return nil
}

func (cl *CacheListener) Debug(m string) error {
	return nil
}

func (cl *CacheListener) Warning(m string) error {
	return nil
}

func (cl *CacheListener) Err(m string) error {
	return nil
}

func (cl *CacheListener) Crit(m string) error {
	return nil
}

func (cl *CacheListener) GetMsgs() []string {
	cl.Lock()
	defer cl.Unlock()
	ms := cl.msgs
	cl.msgs = nil
	return ms
}

// Cache tests have to be run in sequence, and should all call CacheInit() followed by
// cachePurgeAllSync() to clean the cache of any prior data.

// to-do here
//
// - Really if there's a Budget -nn message there should be at least one subsequent Purging message,
//   and if there's a Budget nn message there should be no subsequent Purging (probably).
// - Probably these Infof messages should be under DEBUG and DEBUG should be auto-enabled when
//   running tests (how?)

func TestCaching(t *testing.T) {
	if DEBUG {
		t.Fatal("Hi there")
	}
	d := tmpCopyTree(theClusterDir)
	defer os.RemoveAll(d)

	// 200 is small enough that we should see some purging.
	CacheInit(300)
	cachePurgeAllSync()

	var ul CacheListener
	Log.SetUnderlying(&ul)
	Log.LowerLevelTo(status.LogLevelInfo)

	pc, err := OpenPersistentCluster(d, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Caching and purging test

	// Read five times to stress the cache a bit
	numReads := 5
	for i := 0; i < numReads; i++ {
		_, _, err = pc.ReadSamples(
			time.Date(2023, 05, 01, 0, 0, 0, 0, time.UTC),
			time.Date(2023, 06, 30, 0, 0, 0, 0, time.UTC),
			nil,
			false,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	msgs := ul.GetMsgs()

	filesToCache := []string{
		"/2023/06/05/b.csv",
		"/2023/06/01/a.csv",
		"/2023/06/01/b.csv",
		"/2023/05/28/a.csv",
		"/2023/06/05/a.csv",
		"/2023/06/02/a.csv",
		"/2023/06/04/b.csv",
		"/2023/05/31/b.csv",
		"/2023/06/04/a.csv",
		"/2023/05/31/a.csv",
		"/2023/05/30/b.csv",
		"/2023/06/03/b.csv",
		"/2023/05/30/a.csv",
		"/2023/05/28/b.csv",
	}
	files := strings.Join(filesToCache, "|")
	shouldSee := regexp.MustCompile(
		"^Caching.*(" + files + ")|^Purging .*(" + files + ") .*internal:capacity|^Budget|^Cache hit .*(" + files + ")",
	)

	// Every file in the list above must become cached at one point or another
	//
	// Every file cached must not already be in the cache.
	//
	// Every file purged must be in the cache.
	//
	// Every file hit must be in the cache.
	//
	// There should be numReads hits+reads for every file.
	found := make(map[string]int)
	inCache := make(map[string]bool)
	for _, m := range msgs {
		ms := shouldSee.FindStringSubmatch(m)
		if ms == nil {
			t.Fatal(m)
		}
		if ms[1] != "" {
			found[ms[1]]++
			if inCache[ms[1]] {
				t.Fatalf("Already in cache %s", ms[1])
			}
			inCache[ms[1]] = true
		}
		if ms[2] != "" {
			if !inCache[ms[2]] {
				t.Fatalf("Purge not in cache %s", ms[2])
			}
			delete(inCache, ms[2])
		}
		if ms[3] != "" {
			found[ms[3]]++
			if !inCache[ms[3]] {
				t.Fatalf("Hit not in cache %s", ms[3])
			}
		}
	}
	for _, f := range filesToCache {
		if found[f] != numReads {
			t.Fatal(f)
		}
	}

	// Write causes flush of clean cached data
	//
	//  - clear cache
	//  - read a file that will fit in the cache
	//  - observe Caching msg
	//  - append to it w/o other traffic, see nothing
	//  - read it
	//  - observe Purging b/c internal:dirty
	//  - observe Caching msg
	//  - observe that the contents include the appended data

	cachePurgeAllSync()
	_ = ul.GetMsgs()

	// This should read 2023/05/31/a.csv
	glob, _ := hostglob.NewGlobber(false, []string{"a"})
	_, _, err = pc.ReadSamples(
		time.Date(2023, 05, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
		glob,
		false,
	)

	msgs = ul.GetMsgs()
	if len(msgs) != 1 {
		t.Fatal("Too much action", msgs)
	}
	m, _ := regexp.MatchString("Caching.*a\\.csv", msgs[0])
	if !m {
		t.Fatal("Too much action", msgs)
	}

	err = pc.AppendSamplesAsync("a", "2023-05-31T14:30:38+02:00", "v=0.11.1,time=2023-05-31T14:30:38+02:00,host=a,user=larstha,cmd=awk")
	if err != nil {
		t.Fatal(err)
	}
	sampleBlobs, _, err := pc.ReadSamples(
		time.Date(2023, 05, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
		glob,
		false,
	)
	msgs = ul.GetMsgs()
	if len(msgs) != 2 {
		t.Fatal("Too much action", msgs)
	}
	m, _ = regexp.MatchString("Purging.*a\\.csv.*internal:dirty", msgs[0])
	if !m {
		t.Fatal("Missing purging msg", msgs)
	}
	m, _ = regexp.MatchString("Caching.*a\\.csv", msgs[1])
	if !m {
		t.Fatal("Missing caching msg", msgs)
	}

	// There should be exactly one record matching cmd from above in that record set, as there were
	// no awk commands to begin with

	awks := 0
	for _, s := range uslices.Catenate(sampleBlobs) {
		if s.Cmd.String() == "awk" {
			awks++
		}
	}
	if awks != 1 {
		t.Fatal("append failed")
	}

	// Purging of dirty data causes writeback
	//
	//  - clear cache
	//  - read a file that will fit in the cache
	//  - observe Caching msg
	//  - append to it w/o other traffic, see nothing
	//  - read another file that will cause the cache to overflow
	//  - observe Purging b/c internal:dirty
	//  - observe Caching msg of the new file
	//  - Read the old file again
	//  - observe that the contents include the appended data

	cachePurgeAllSync()
	_ = ul.GetMsgs()

	_, _, err = pc.ReadSamples(
		time.Date(2023, 05, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
		glob,
		false,
	)

	msgs = ul.GetMsgs()
	if len(msgs) != 1 {
		t.Fatal("Too much action", msgs)
	}
	m, _ = regexp.MatchString("Caching.*a\\.csv", msgs[0])
	if !m {
		t.Fatal("Too much action", msgs)
	}

	err = pc.AppendSamplesAsync("a", "2023-05-31T14:30:38+02:00", "v=0.11.1,time=2023-05-31T14:30:38+02:00,host=a,user=larstha,cmd=zappa")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = pc.ReadSamples(
		time.Date(2023, 05, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 05, 29, 0, 0, 0, 0, time.UTC),
		glob,
		false,
	)

	msgs = ul.GetMsgs()
	if len(msgs) != 2 {
		t.Fatal("Too much action", msgs)
	}
	m, _ = regexp.MatchString("Purging.*05/31/a\\.csv.*internal:dirty", msgs[0])
	if !m {
		t.Fatal("Missing purging msg", msgs)
	}
	m, _ = regexp.MatchString("Caching.*05/28/.*\\.csv", msgs[1])
	if !m {
		t.Fatal("Missing caching msg", msgs)
	}
	sampleBlobs, _, err = pc.ReadSamples(
		time.Date(2023, 05, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 06, 01, 0, 0, 0, 0, time.UTC),
		glob,
		false,
	)
	zappas := 0
	for _, s := range uslices.Catenate(sampleBlobs) {
		if s.Cmd.String() == "zappa" {
			zappas++
		}
	}
	if zappas != 1 {
		t.Fatal("append failed")
	}
}
