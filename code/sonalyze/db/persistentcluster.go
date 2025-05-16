// A PersistentCluster is backed by a sonarlog date-indexed data store and will find and manage
// files in that store.
//
//
// Overall.
//
// The PersistentCluster maintains a shadow directory tree for a date range [From, To] (represented
// as an inclusive "from" and and exclusive "to" date, both set to midnight of their respective
// dates).  Every subdirectory with a name of the form yyyy/mm/dd in that date range is known to the
// shadow tree.  When presented with a new from,to range the first thing we do is make sure the
// shadow tree is populated with the directories.
//
// In each (leaf) directory there are lists of files matching specific semantic patterns, documented
// in the block comment "FILE NAME SCHEMES" in db/clusterstore.go.  These lists are populated lazily
// when the files in the directory are first needed.
//
// A specific file can be requested for a given date when inserting data.  The subdirectory for that
// date may not exist and will first have to be created.  Then the file may not exist and may also
// have to be created.
//
// There is an invariant that if we know any files in a leaf directory then we know all the files in
// that directory: we only need to scan the directory once, and if a new file is created it is
// created via the PersistentCluster, not behind our back.  (It may be possible to be a little
// resilient about this by using NoCreate flags to file operations and then handling errors, but
// I've not worried about this.)
//
// File data may (eventually) be cached, but that is transparent to the PersistentCluster, it is
// handled at the level of the LogFile.
//
//
// Memory use.
//
// Over time (months) as the server is up, there may be a substantial amount of metadata in memory -
// the shadow directory trees may become quite large.  For example, most queries will only very
// rarely request files from more than a few days ago and caching really old stuff may be
// undesirable.  In principle we would want to purge disused shadow directory tree entries.
// However, this is hard because we require that there is only ever one LogFile per appendable file,
// and so we need to prove, before we purge a directory, that no references exist anywhere in the
// system to any file in the directory (or we risk the directory becoming reinstated and a new
// LogFile created for one that is already active).  In addition, given the constraints of the date
// range, we'd only ever be able to purge from the ends of the date range.
//
// On the other hand, let's consider how much data we're talking about.  For one year there are 365
// directories, not an important factor.  But for saga+betzy+fram+fox there are > 2000 nodes, and
// one file for each per day.  Even if we intern all the strings (and we probably should) then each
// file structure is at least 64 bytes, and not pointer-free.  2000*365*64=46MB.  I guess it's
// probably affordable, for now.
//
// A halfway solution, should non-purging turn out to be a problem, is to purge *everything* every
// so often, effectively restarting the daemon.

package db

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"go-utils/config"
	"go-utils/hostglob"
	uslices "go-utils/slices"
	. "sonalyze/common"
)

// Locking discipline (notes).
//
// In many contexts, the mutex on the PersistentCluster dominates the mutexes on the files in the
// cluster.  However, files must be locked individually by all mutable operations.  In a number of
// cases, file operations may be performed without the cluster lock being held.  The file should not
// assume that the cluster lock is held (and should ideally not know the cluster it belongs to).
//
// Methods that assume the lock is held on self when called are named whateverMethodLocked().

type PersistentCluster struct /* implements AppendableCluster */ {
	// MT: Immutable after initialization
	sampleFiles  sampleFilesAdapter
	sysinfoFiles sysinfoFilesAdapter
	sacctFiles   sacctFilesAdapter
	cluzterFiles cluzterFilesAdapter

	// MT: Immutable after initialization
	samplesMethods           ReadSyncMethods
	loadDataMethods          ReadSyncMethods
	gpuDataMethods           ReadSyncMethods
	nodeConfigRecordsMethods ReadSyncMethods
	sacctMethods             ReadSyncMethods
	cluzterMethods           ReadSyncMethods

	// MT: Immutable after initialization
	// The dataDir must have been path.Clean'd, it is the root directory for the cluster.
	dataDir string

	// MT: Immutable after initialization
	cfg *config.ClusterConfig

	sync.Mutex
	closed bool

	// The shadow directory tree underneath dataDir.  The timestamps are 00:00:00 UTC the earliest
	// day for which we have created directories and the start of the day after the latest day,
	// respectively.  There may be directories outside this range.  For each directory, its file
	// list may be empty - directories may be scanned lazily.  For a date in the scanned range, all
	// subdirectories are in this map - there can be no subdirectories created outside the system.
	// This list is sorted lexicographically by the `name` field, ie, ascending by date.
	dirs     []*persistentDir
	fromDate time.Time
	toDate   time.Time

	// A set of files that have had data appended but have not yet been flushed.  Flushing is
	// triggered by FlushAsync().
	dirty map[*LogFile]bool
}

type persistentDir struct {
	// Path name underneath the cluster's dataDir, form "yyyy/mm/dd".
	name string

	// All files in the directory.  If files is nil then the directory has not been scanned for
	// those files, otherwise each map is canonical - there can be no files (of that type) in the
	// directory that are not in this map.  There is a separate set per file because these are
	// independent.  We could have represented these as a single map indexed by the glob.
	sampleFiles  map[string]*LogFile
	sysinfoFiles map[string]*LogFile
	sacctFiles   map[string]*LogFile
	cluzterFiles map[string]*LogFile
}

func newPersistentCluster(dataDir string, cfg *config.ClusterConfig) *PersistentCluster {
	// Initially, populate for today's date.
	fromDate := ThisDay(time.Now().UTC())
	toDate := fromDate.AddDate(0, 0, 1)
	dirs := findSortedDateIndexedDirectories(dataDir, fromDate, toDate)
	return &PersistentCluster{
		samplesMethods:           newSampleFileMethods(cfg, sampleFileKindSample),
		loadDataMethods:          newSampleFileMethods(cfg, sampleFileKindLoadDatum),
		gpuDataMethods:           newSampleFileMethods(cfg, sampleFileKindGpuDatum),
		nodeConfigRecordsMethods: newSysinfoFileMethods(cfg),
		sacctMethods:             newSacctFileMethods(cfg),
		cluzterMethods:           newCluzterFileMethods(cfg),
		dataDir:                  dataDir,
		cfg:                      cfg,
		dirs:                     dirs,
		fromDate:                 fromDate,
		toDate:                   toDate,
		dirty:                    make(map[*LogFile]bool),
	}
}

func (pc *PersistentCluster) Config() *config.ClusterConfig {
	pc.Lock()
	defer pc.Unlock()
	if pc.closed {
		return nil
	}

	return pc.cfg
}

func (pc *PersistentCluster) Close() error {
	pc.Lock()
	defer pc.Unlock()
	if pc.closed {
		return ClusterClosedErr
	}

	pc.closed = true

	pc.flushSyncLocked()

	// It's not technically necessary to purge the cache since these files, belonging to a cluster
	// no longer in memory, will be purged eventually anyway, but it does free up memory more
	// quickly and will improve the hit rate for other clusters - assuming clusters are closed
	// individually during the run and not just at the end of the run.
	for _, d := range pc.dirs {
		for _, f := range d.sampleFiles {
			f.PurgeCache("closing cluster")
		}
		for _, f := range d.sysinfoFiles {
			f.PurgeCache("closing cluster")
		}
	}

	return nil
}

func (pc *PersistentCluster) FlushAsync() {
	pc.Lock()
	defer pc.Unlock()
	if pc.closed {
		return
	}

	// TODO: IMPROVEME.  Since we're supposed to trigger async flushing, make this async.
	pc.flushSyncLocked()
}

// Pre: LOCK HELD
func (pc *PersistentCluster) flushSyncLocked() {
	for file := range pc.dirty {
		file.FlushSync()
	}
	clear(pc.dirty)
}

// Return cleaned file names that will be passed to os.Open().  The slice should be considered
// immutable.  In particular, it should not be sorted.
//
// No caching of the enumerated file list is done here at the moment b/c once we implement caching
// of files, directory enumeration will happen only once and then a different structure will be
// built from that.

func (pc *PersistentCluster) SampleFilenames(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
) ([]string, error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return pc.findFilenames(fromDate, toDate, hosts, &pc.sampleFiles)
}

func (pc *PersistentCluster) SysinfoFilenames(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
) ([]string, error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return pc.findFilenames(fromDate, toDate, hosts, &pc.sysinfoFiles)
}

func (pc *PersistentCluster) SacctFilenames(
	fromDate, toDate time.Time,
) ([]string, error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return pc.findFilenames(fromDate, toDate, nil, &pc.sacctFiles)
}

func (pc *PersistentCluster) CluzterFilenames(
	fromDate, toDate time.Time,
) ([]string, error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return pc.findFilenames(fromDate, toDate, nil, &pc.cluzterFiles)
}

func (pc *PersistentCluster) findFilenames(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	fa filesAdapter,
) ([]string, error) {
	pc.Lock()
	defer pc.Unlock()
	if pc.closed {
		return nil, ClusterClosedErr
	}

	files := pc.findFilesLocked(fromDate, toDate, hosts, fa)
	return filenames(files), nil
}

func (pc *PersistentCluster) ReadSamples(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	verbose bool,
) (sampleBlobs [][]*Sample, dropped int, err error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return readPersistentClusterRecords(
		pc, fromDate, toDate, hosts, verbose, &pc.sampleFiles, pc.samplesMethods,
		readSampleSlice,
	)
}

func (pc *PersistentCluster) ReadLoadData(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	verbose bool,
) (dataBlobs [][]*LoadDatum, dropped int, err error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return readPersistentClusterRecords(
		pc, fromDate, toDate, hosts, verbose, &pc.sampleFiles, pc.loadDataMethods,
		readLoadDatumSlice,
	)
}

func (pc *PersistentCluster) ReadGpuData(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	verbose bool,
) (dataBlobs [][]*GpuDatum, dropped int, err error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return readPersistentClusterRecords(
		pc, fromDate, toDate, hosts, verbose, &pc.sampleFiles, pc.gpuDataMethods,
		readGpuDatumSlice,
	)
}

func (pc *PersistentCluster) ReadSysinfoData(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	verbose bool,
) (sysinfoBlobs [][]*config.NodeConfigRecord, dropped int, err error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return readPersistentClusterRecords(
		pc, fromDate, toDate, hosts, verbose, &pc.sysinfoFiles, pc.nodeConfigRecordsMethods,
		readNodeConfigRecordSlice,
	)
}

func (pc *PersistentCluster) ReadSacctData(
	fromDate, toDate time.Time,
	verbose bool,
) (sacctBlobs [][]*SacctInfo, dropped int, err error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return readPersistentClusterRecords(
		pc, fromDate, toDate, nil, verbose, &pc.sacctFiles, pc.sacctMethods,
		readSacctSlice,
	)
}

func (pc *PersistentCluster) ReadCluzterData(
	fromDate, toDate time.Time,
	verbose bool,
) (cluzterBlobs [][]*CluzterInfo, dropped int, err error) {
	if DEBUG {
		Assert(fromDate.Location() == time.UTC, "UTC expected")
		Assert(toDate.Location() == time.UTC, "UTC expected")
	}
	return readPersistentClusterRecords(
		pc, fromDate, toDate, nil, verbose, &pc.cluzterFiles, pc.cluzterMethods,
		readCluzterSlice,
	)
}

func readPersistentClusterRecords[V any, U ~[][]*V](
	pc *PersistentCluster,
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	verbose bool,
	fa filesAdapter,
	methods ReadSyncMethods,
	reader func(files []*LogFile, verbose bool, methods ReadSyncMethods) (U, int, error),
) (recordBlobs U, dropped int, err error) {
	// TODO: IMPROVEME: Don't hold the lock while reading, it's not necessary, caching is per-file
	// and does not interact with the cluster.  But be sure to get pc.cfg while holding the lock.
	pc.Lock()
	defer pc.Unlock()
	if pc.closed {
		return nil, 0, ClusterClosedErr
	}

	files := pc.findFilesLocked(fromDate, toDate, hosts, fa)
	return reader(files, verbose, methods)
}

// For file name patterns, see the comment "FILE NAME SCHEMES" in clusterstore.go.

func (pc *PersistentCluster) AppendSamplesAsync(ty FileAttr, host, timestamp string, payload any) error {
	switch ty {
	case FileSampleCSV:
		return pc.appendDataAsync(timestamp, fmt.Sprintf("%s.csv", host), payload, pc.sampleFiles)
	case FileSampleV0JSON:
		return pc.appendDataAsync(timestamp, fmt.Sprintf("0+sample-%s.json", host), payload, pc.sampleFiles)
	default:
		panic("Unsupported 'sample' data format")
	}
}

func (pc *PersistentCluster) AppendSysinfoAsync(ty FileAttr, host, timestamp string, payload any) error {
	switch ty {
	case FileSysinfoOldJSON:
		return pc.appendDataAsync(
			timestamp, fmt.Sprintf("sysinfo-%s.json", host), payload, pc.sysinfoFiles)
	case FileSysinfoV0JSON:
		return pc.appendDataAsync(
			timestamp, fmt.Sprintf("0+sysinfo-%s.json", host), payload, pc.sysinfoFiles)
	default:
		panic("Unsupported 'sysinfo' data format")
	}
}

func (pc *PersistentCluster) AppendSlurmSacctAsync(ty FileAttr, timestamp string, payload any) error {
	switch ty {
	case FileSlurmCSV:
		return pc.appendDataAsync(timestamp, "slurm-sacct.csv", payload, pc.sacctFiles)
	case FileSlurmV0JSON:
		return pc.appendDataAsync(timestamp, "0+job-slurm.json", payload, pc.sacctFiles)
	default:
		panic("Unsupported 'slurm' data format")
	}
}

func (pc *PersistentCluster) AppendCluzterAsync(ty FileAttr, timestamp string, payload any) error {
	switch ty {
	case FileCluzterV0JSON:
		return pc.appendDataAsync(timestamp, "0+cluzter-slurm.json", payload, pc.cluzterFiles)
	default:
		panic("Unsupported 'cluzter' data format")
	}
}

func (pc *PersistentCluster) appendDataAsync(
	timestamp,
	filename string,
	payload any,
	fa filesAdapter,
) error {
	// A little hair so as not to hold the cluster lock while appending to the file.  The pattern
	// can be abstracted if we end up using it elsewhere.
	pc.Lock()
	shouldUnlock := true
	defer func() {
		if shouldUnlock {
			pc.Unlock()
		}
	}()
	if pc.closed {
		return ClusterClosedErr
	}

	file, err := pc.findFileByTimeLocked(timestamp, filename, fa)
	if err != nil {
		return err
	}

	pc.dirty[file] = true

	shouldUnlock = false
	pc.Unlock()
	return file.AppendAsync(payload)
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Adapters to hide file idiosyncracies in a persistent store.

type filesAdapter interface {
	globs() []string
	proscribedBasename(basename string) bool
	fileTypeFromBasename(basename string) FileAttr
	getFiles(*persistentDir) map[string]*LogFile
	setFiles(*persistentDir, map[string]*LogFile)
}

type sampleFilesAdapter struct {
}

func (_ sampleFilesAdapter) globs() []string {
	return []string{"*.csv", "0+sample-*.json"}
}

func (_ sampleFilesAdapter) getFiles(d *persistentDir) map[string]*LogFile {
	return d.sampleFiles
}

func (_ sampleFilesAdapter) setFiles(d *persistentDir, files map[string]*LogFile) {
	d.sampleFiles = files
}

func (_ sampleFilesAdapter) proscribedBasename(fn string) bool {
	// Backward compatible: remove cpuhog.csv and bughunt.csv, these later moved into separate
	// directory trees.
	//
	// TODO: REMOVEME: Remove those tests once we know that all data stores have been cleaned of
	// those files.
	return fn == "cpuhog.csv" || fn == "bughunt.csv" || fn == "slurm-sacct.csv"
}

func (_ sampleFilesAdapter) fileTypeFromBasename(basename string) FileAttr {
	if strings.HasPrefix(basename, "0+sample-") {
		return FileSampleV0JSON
	}
	return FileSampleCSV
}

type sysinfoFilesAdapter struct {
}

func (_ sysinfoFilesAdapter) globs() []string {
	return []string{"sysinfo-*.json", "0+sysinfo-*.json"}
}

func (_ sysinfoFilesAdapter) getFiles(d *persistentDir) map[string]*LogFile {
	return d.sysinfoFiles
}

func (_ sysinfoFilesAdapter) setFiles(d *persistentDir, files map[string]*LogFile) {
	d.sysinfoFiles = files
}

func (_ sysinfoFilesAdapter) proscribedBasename(fn string) bool {
	return false
}

func (_ sysinfoFilesAdapter) fileTypeFromBasename(basename string) FileAttr {
	if strings.HasPrefix(basename, "0+sysinfo-") {
		return FileSysinfoV0JSON
	}
	return FileSysinfoOldJSON
}

type sacctFilesAdapter struct {
}

func (_ sacctFilesAdapter) globs() []string {
	return []string{"slurm-sacct.csv", "0+job-slurm.json"}
}

func (_ sacctFilesAdapter) getFiles(d *persistentDir) map[string]*LogFile {
	return d.sacctFiles
}

func (_ sacctFilesAdapter) setFiles(d *persistentDir, files map[string]*LogFile) {
	d.sacctFiles = files
}

func (_ sacctFilesAdapter) proscribedBasename(fn string) bool {
	return false
}

func (_ sacctFilesAdapter) fileTypeFromBasename(basename string) FileAttr {
	if basename == "slurm-sacct.csv" {
		return FileSlurmCSV
	}
	return FileSlurmV0JSON
}

type cluzterFilesAdapter struct {
}

func (_ cluzterFilesAdapter) globs() []string {
	return []string{"0+cluzter-slurm.json"}
}

func (_ cluzterFilesAdapter) getFiles(d *persistentDir) map[string]*LogFile {
	return d.cluzterFiles
}

func (_ cluzterFilesAdapter) setFiles(d *persistentDir, files map[string]*LogFile) {
	d.cluzterFiles = files
}

func (_ cluzterFilesAdapter) proscribedBasename(fn string) bool {
	return false
}

func (_ cluzterFilesAdapter) fileTypeFromBasename(basename string) FileAttr {
	return FileCluzterV0JSON
}

///////////////////////////////////////////////////////////////////////////////////////////////////
//
// Low-level stuff.

func (pc *PersistentCluster) findFilesLocked(
	fromDate, toDate time.Time,
	hosts *hostglob.HostGlobber,
	fa filesAdapter,
) []*LogFile {
	globs := fa.globs()

	// Find all matching files in the date range.
	fromDate = ThisDay(fromDate)
	toDate = RoundupDay(toDate)

	files := make([]*LogFile, 0)
	for _, d := range pc.selectDirsLocked(fromDate, toDate) {
		if fa.getFiles(d) == nil {
			basenames := make([]string, 0)
			for _, glob := range globs {
				basenames = append(basenames, findFiles(pc.dataDir, d.name, glob)...)
			}

			// Filter file names that are simply disallowed
			{
				dest := 0
				for src := 0; src < len(basenames); src++ {
					fn := basenames[src]
					if fa.proscribedBasename(fn) {
						continue
					}
					basenames[dest] = basenames[src]
					dest++
				}
				basenames = basenames[:dest]
			}

			newFiles := make(map[string]*LogFile, len(basenames))
			for _, name := range basenames {
				f := newLogFile(
					Fullname{
						cluster:  pc.dataDir,
						dirname:  d.name,
						basename: name,
					},
					fileAppendable|fa.fileTypeFromBasename(name),
				)
				newFiles[name] = f
			}
			fa.setFiles(d, newFiles)
		}

		// Retain only files whose names match the filter, if present
		if hosts != nil && !hosts.IsEmpty() {
			for _, glob := range globs {
				extensionLen := len(glob) - strings.LastIndexByte(glob, '.')
				for _, c := range fa.getFiles(d) {
					fn := c.basename
					if hosts.Match(fn[:len(fn)-extensionLen]) {
						files = append(files, c)
					}
				}
			}
		} else {
			for _, c := range fa.getFiles(d) {
				files = append(files, c)
			}
		}
	}

	return files
}

// Pre: LOCK HELD
// Pre: !ld.closed
// Post if !error: directory exits on disk
// Post if !error: directory exists in cluster's tree
func (pc *PersistentCluster) findFileByTimeLocked(
	timestamp, filename string,
	fa filesAdapter,
) (*LogFile, error) {
	tval, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return nil, BadTimestampErr
	}
	// This conversion to UTC creates a detectable semantic change from the Rust code and earlier Go
	// code.  It means that a record for eg 2024-06-03T00:00:01+02:00 will end up in the 2024/06/02
	// bin as the date is normalized to 2024-06-02T22:00:01Z.  This is arguably the right thing, as
	// making every date UTC following ingest will keep things relatively sane.  But it means that
	// it will no longer be easy to compare log files across this change.
	tval = tval.UTC()
	d, err := pc.ensureScannedDirectoryLocked(tval)
	if err != nil {
		return nil, err
	}
	var f *LogFile
	name := Fullname{
		cluster:  pc.dataDir,
		dirname:  d.name,
		basename: filename,
	}
	attrs := fileAppendable | fa.fileTypeFromBasename(filename)
	files := fa.getFiles(d)
	if files == nil {
		files = make(map[string]*LogFile, 5)
		fa.setFiles(d, files)
	}
	f = files[filename]
	if f == nil {
		f = newLogFile(name, attrs)
		files[filename] = f
	}
	return f, nil
}

// Return the subslice of the shadown directory slice corresponding to the date range.  For correct
// results, `fromDate` and `toDate` must be rounded to midnight and should be UTC.

func (pc *PersistentCluster) selectDirsLocked(fromDate, toDate time.Time) []*persistentDir {
	pc.ensureScannedDirectoriesLocked(fromDate, toDate)
	fromLoc, _ := binarySearchDirs(pc.dirs, fromDate)
	toLoc, _ := binarySearchDirs(pc.dirs, toDate)
	return pc.dirs[fromLoc:toLoc]
}

// Make sure a directory exists for the given time, and return it.  The time should be UTC.

func (pc *PersistentCluster) ensureScannedDirectoryLocked(t time.Time) (*persistentDir, error) {
	fromDate := ThisDay(t)
	toDate := fromDate.AddDate(0, 0, 1)
	pc.ensureScannedDirectoriesLocked(fromDate, toDate)

	// Find or create the directory
	ix, found := binarySearchDirs(pc.dirs, fromDate)
	if !found {
		name := dirnameFromTime(t)
		err := os.MkdirAll(path.Join(pc.dataDir, name), dirPermissions)
		if err != nil {
			return nil, err
		}
		d := &persistentDir{
			name: name,
		}
		// Insert it.  We don't need to update pc.fromDate/toDate because the directories have been
		// scanned to at least that range.
		pc.dirs = slices.Insert(pc.dirs, ix, d)
	}
	return pc.dirs[ix], nil
}

// Make sure the shadow directory tree includes all matching subdirectories from `fromDate`
// (inclusive) to `toDate` (exclusive).  For correct results, `fromDate` and `toDate` must be
// rounded to midnight, and should be UTC.

func (pc *PersistentCluster) ensureScannedDirectoriesLocked(fromDate, toDate time.Time) {
	var prefix, suffix []*persistentDir
	if fromDate.Before(pc.fromDate) {
		prefix = findSortedDateIndexedDirectories(pc.dataDir, fromDate, pc.fromDate)
		pc.fromDate = fromDate
	}

	if pc.toDate.Before(toDate) {
		suffix = findSortedDateIndexedDirectories(pc.dataDir, pc.toDate, toDate)
		pc.toDate = toDate
	}

	if prefix == nil && suffix != nil {
		pc.dirs = append(pc.dirs, suffix...)
	} else if prefix != nil {
		result := make([]*persistentDir, len(prefix)+len(pc.dirs)+len(suffix))
		copy(result, prefix)
		copy(result[len(prefix):], pc.dirs)
		copy(result[len(prefix)+len(pc.dirs):], suffix)
		pc.dirs = result
	}

	if DEBUG {
		// Assert that the list is strictly ascending.
		for i := 1; i < len(pc.dirs); i++ {
			if pc.dirs[i-1].name >= pc.dirs[i].name {
				panic(fmt.Sprintf("Out of order: %s %s", pc.dirs[i-1].name, pc.dirs[i].name))
			}
		}
	}
}

// Scan the directory for files of the given kind and return the matches.

func findFiles(dataDir, dirname, pattern string) []string {
	matches, _ := fs.Glob(os.DirFS(dataDir), path.Join(dirname, pattern))
	return uslices.Map(matches, func(s string) string { return path.Base(s) })
}

// Scan the tree underneath `dataDir` for subdirectories named yyyy/mm/dd for the date range from
// `from` (inclusive) to `to` (exclusive) and return a list of new persistentDir sorted ascending by
// directory name.  For correct results, `from` and `to` must be rounded to midnight.  Errors are
// ignored, as are non-directories with names that match the pattern.
//
// For best results, from and to should be UTC.

func findSortedDateIndexedDirectories(dataDir string, from, to time.Time) []*persistentDir {
	filesys := os.DirFS(dataDir).(fs.StatFS)
	result := []*persistentDir{}
	for ; from.Before(to); from = from.AddDate(0, 0, 1) {
		probeFn := dirnameFromTime(from)
		info, err := filesys.Stat(probeFn)
		if err != nil || !info.IsDir() {
			continue
		}
		result = append(result, &persistentDir{name: probeFn})
	}
	return result
}

// Return the index of the directory with name d, or the index of the record s.t. d would come
// before that record.
//
// Note, this depends on d.name not changing, or we'll have a race.  Normally this is called with
// the lock held and it's not a problem anyway.
//
// For best results, d should be UTC.

func binarySearchDirs(dirs []*persistentDir, d time.Time) (int, bool) {
	return slices.BinarySearchFunc(dirs, dirnameFromTime(d), func(d *persistentDir, s string) int {
		if d.name == s {
			return 0
		}
		if d.name < s {
			return -1
		}
		return 1
	})
}

func dirnameFromTime(t time.Time) string {
	return fmt.Sprintf("%04d/%02d/%02d", t.Year(), t.Month(), t.Day())
}
