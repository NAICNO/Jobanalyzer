// LogFile content caching logic.
//
// A LogFile can choose to cache its content (in whatever form it wants).
//
// - To cache some data with a recorded size, call lf.cacheWriteLocked()
// - To obtain the data (if still there), call lf.cacheReadLocked()
// - To purge the data, call lf.cachePurgeLocked()
// - To check whether there are data, call lf.isCachedLocked()
//
// Behind the scenes, there is a cache budget, tracking the amount of data available before the
// cache is full.  Should a write to the cache bring the cache budget below zero, the cache will
// schedule a background purge.  The purge will pick an item (a LogFile with cached data) to purge
// from all the files that have cached data, and then purge it by clearing its cache.  It will
// repeat this until the cache budget is nonnegative.
//
// Note that since the purging is asynchronous, it is never possible to know the cached state of a
// file unless the lock is held.
//
// The LogFile.whateverLocked() methods must be called with the file's lock held.
//
// Additionally, there is a data structure for managing background purging that has a separate lock,
// the purgeLock.  The fields of the LogFile that participate in purging are covered *also* by the
// purgeLock.  Asynchronous purging holds only the purgeLock, all other operations hold first the
// file lock and then the purgeLock.
//
// For an overview of locking, see clusterstore.go.

package filedb

import (
	"math/rand"
	"os"
	"sync"
	"sync/atomic"

	. "sonalyze/common"
)

// Cache data to be embedded in a LogFile, protected by the LogFile mutex, accessed exclusively
// by the lf.whateverLocked() methods below.
type logFileCacheData struct {
	isCached  bool  // true if cache is populated
	cacheData any   // eg *sampleCachePayload
	cacheSize int64 // estimated size in bytes of cached content
}

// More cache data for embedding, but these are proteced by the purgeLock and accessible only to the
// underflow manager (with or without the LogFile lock held).
type logFilePurgeableData struct {
	cacheLru uint64 // updated when cache is populated or read
	cacheIx  int    // index in purgeable table
}

const (
	// Underflow message queue length.  Hard to know what size is good, 100 is possibly overkill,
	// but we don't want senders of this to be stalling.
	//
	// TODO: IMPROVEME: A queue may be the wrong abstraction anyway - what we want is a signal to
	// wake up the cleaner, if it is not already awake.  Maybe a sync.Cond would be better.  We
	// would want a dedicated lock for that, not share the purgeLock.
	cacheUnderflowCap = 100
)

var (
	// MT: Atomic
	enabled atomic.Bool  // Cache is in use
	budget  atomic.Int64 // Signed b/c we need this to be able to go negative on overflow

	// MT: Constant after initialization; thread-safe
	cacheUnderflow = make(chan bool, cacheUnderflowCap)
)

// Data to manage purging.
//
// purgeable is the set of files that can be purged from the cache.  The index of a file in the set
// is in the file's cacheIx member.  In general, a file has cached data iff it is a member of
// purgeable.
//
// TODO: OPTIMIZEME: The purgeable array can become large, in the hundreds of thousands of elements.
// It's possible that a log-structure would be a better solution.

var (
	// MT: Locked
	purgeLock   sync.Mutex
	purgeable   = make([]*LogFile, 0, 10000)
	lruCounter  uint64
	lruOverflow bool // Set to true if the counter overflows
)

func CacheInit(cacheSize int64) {
	if cacheSize < 0 {
		Log.Infof("Disabling cache")
		enabled.Store(false)
		return
	}

	Log.Infof("Enabling cache, initial budget %d", cacheSize)
	budget.Store(cacheSize)
	enabled.Store(true)
}

// This is for testing
func CachePurgeAllSync() {
	for {
		f := pickFileToPurge()
		if f == nil {
			break
		}
		f.PurgeCache("testing")
	}
}

// Return true iff caching is enabled.  This will not block.
func CacheEnabled() bool {
	return enabled.Load()
}

func (lf *LogFile) isCachedLocked() bool {
	return lf.isCached
}

func (lf *LogFile) cachePurgeLocked(reason string) {
	if !lf.isCached {
		return
	}

	Log.Infof("Purging %s b/c %s", lf.Fullname, reason)

	lf.removeFromPurgeableLocked()

	budget.Add(lf.cacheSize)

	lf.isCached = false
	lf.cacheSize = 0
	lf.cacheData = nil
}

func (lf *LogFile) cacheReadLocked() (bool, any) {
	if !lf.isCached {
		return false, nil
	}
	Log.Infof("Cache hit %s", lf.Fullname)
	return true, lf.cacheData
}

func (lf *LogFile) cacheWriteLocked(data any, size int64) {
	if lf.isCached {
		lf.cachePurgeLocked("internal:replacing")
	}

	Log.Infof("Caching %s size %d", lf.Fullname, size)

	lf.isCached = true
	lf.cacheSize = size
	lf.cacheData = data

	lf.addToPurgeableLocked()

	if budget.Add(-size) < 0 {
		cacheUnderflow <- true
	}
}

func (lf *LogFile) removeFromPurgeableLocked() {
	purgeLock.Lock()
	defer purgeLock.Unlock()

	x := purgeable[len(purgeable)-1]
	x.cacheIx = lf.cacheIx
	purgeable[lf.cacheIx] = x
	purgeable = purgeable[:len(purgeable)-1]
	lf.cacheLru = 0
	lf.cacheIx = -1
}

func (lf *LogFile) addToPurgeableLocked() {
	purgeLock.Lock()
	defer purgeLock.Unlock()

	lf.cacheIx = len(purgeable)
	lf.cacheLru = lruCounter
	lruCounter++
	if lruCounter == 0 {
		lruOverflow = true
	}

	purgeable = append(purgeable, lf)
}

// A goroutine that performs cache clearing asynchronously.  No locks are held while the outer loop
// of this routine runs.  It will take file locks and purgeLock as necessary.

func cacheUnderflowFunc() {
	<-cacheUnderflow
	for {
		b := budget.Load()
		Log.Infof("Budget %d", b)
		if b >= 0 {
			break
		}
		lf := pickFileToPurge()
		if lf == nil {
			break
		}
		lf.PurgeCache("internal:capacity")
		// Discard pending messages because we'll go again and we don't want the queue to back up.
		// See comments above about using a sync.Cond instead of a queue for this signalling.
	again:
		select {
		case dummy := <-cacheUnderflow:
			_ = dummy
			goto again
		default:
		}
	}
}

func init() {
	go Forever(cacheUnderflowFunc, os.Stderr)
}

// This uses 2-random LRU for the selection: pick two items at random and purge the LRU of them.
// See https://danluu.com/2choices-eviction/ and works referenced at the end of that post.
//
// (Full LRU would win if we had perfect information (we do) very cheaply (we'd need at least a
// priority queue, with its O(log n) extraction cost, and log n will be in the range 17-18 here),
// and 2-random LRU beats plain random by a good margin.  And 2-random is simple.)

func pickFileToPurge() *LogFile {
	purgeLock.Lock()
	defer purgeLock.Unlock()

	// Overflow happens only extremely rarely.  Setting everyone's LRU to zero will turn the
	// selection below into random selection for some time, degrading cache performance, but hot
	// files will get true LRU numbers very quickly and for the rest of the files it doesn't matter.
	if lruOverflow {
		for _, lf := range purgeable {
			lf.cacheLru = 0
		}
		lruOverflow = false
	}

	len := len(purgeable)
	if len == 0 {
		return nil
	}

	a := purgeable[rand.Intn(len)]
	b := purgeable[rand.Intn(len)]

	// Greater LRU means more recently used.
	if a.cacheLru <= b.cacheLru {
		return a
	}
	return b
}
