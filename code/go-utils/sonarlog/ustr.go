// Interned string type, this is called "Ustr" because that's what it's called in the Rust code.
//
// There is a global, thread-safe store of Ustr values.  Casual uses can use StringToUstr() to map a
// string to its Ustr.  However, in multi-threaded situations with a lot of string creation the
// store can become contended.  In this case, use UstrCache (see further down) as a thread-local
// cache.
//
// Facts about Ustr:
//
// - Ustr is 4 bytes and pointer-free
// - StringToUstr("") == 0
// - For a given s, StringToUstr(s) == StringToUstr(s)
// - For distinct s1 != s2, StringToUstr(s1) != StringToUstr(s2)
// - Ustr itself can NOT be compared with "<" et al
// - However we can compare u1.String() vs u2.String() with "<" et al
// - StringToUstr(s).String() is not necessarily the same object as s
// - If u=StringToUstr(s) then u.String() === u.String() (same object)

package sonarlog

import (
	"sync"
)

type Ustr uint32

var (
	tableLock   sync.RWMutex
	internTable map[string]Ustr
	revTable    []string
)

// The zero value of Ustr is the empty string

const UstrEmpty Ustr = 0 // StringToUstr("")

func init() {
	internTable = map[string]Ustr{"": UstrEmpty}
	revTable = []string{""}
}

// Return the Ustr for the string.  This is guaranteed not to retain s.

func StringToUstr(s string) Ustr {
	u, _ := stringToUstrAndString(s)
	return u
}

func (u Ustr) String() string {
	tableLock.RLock()
	defer tableLock.RUnlock()

	return revTable[u]
}

func UstrStats(printStrings bool) {
	tableLock.RLock()
	defer tableLock.RUnlock()

	println(len(internTable))
	if printStrings {
		for _, v := range revTable {
			println(v)
		}
	}
}

// This returns the Ustr and a string for s that is its own object.  This is guaranteed not to
// retain s.

func stringToUstrAndString(s string) (Ustr, string) {
	// The "read" section can be heavily contended even though it should be quick, possibly the
	// string hashing operation is somewhat expensive.
	tableLock.RLock()
	probe, ok := internTable[s]
	if ok {
		s2 := revTable[probe]
		tableLock.RUnlock()
		return probe, s2
	}
	tableLock.RUnlock()

	// But there is a liability in that the string hashing and allocation of the new string (and
	// maybe the revTable extension) occur within the write section, and that creates contention.
	// Depending on how that's all implemented the critical sections are fairly heavy.

	tableLock.Lock()
	defer tableLock.Unlock()

	// There is a window between RUnlock and Lock when the table could have changed so check again.
	if probe, ok := internTable[s]; ok {
		return probe, revTable[probe]
	}

	// Make a copy of the string.  Hoisting these out of the critical section is weirdly not a win.
	b := make([]byte, len(s))
	copy(b, s)
	ns := string(b)

	n := Ustr(len(revTable))
	internTable[ns] = n
	revTable = append(revTable, ns)
	return n, ns
}

type UstrAllocator interface {
	Alloc(s string) Ustr
}

// This is an unsynchronized cache that is a facade for the global Ustr store.  Use this in a
// thread-local way if the global store is very contended.

type UstrCache struct {
	m map[string]Ustr
}

func NewUstrCache() *UstrCache {
	return &UstrCache{m: make(map[string]Ustr)}
}

func (uc *UstrCache) Alloc(s string) Ustr {
	if probe, ok := uc.m[s]; ok {
		return probe
	}
	u, s2 := stringToUstrAndString(s)
	uc.m[s2] = u
	return u
}

// This is just an entry point into the global Ustr store, with the same API as the cache above.

type UstrFacade struct {
	dummy int
}

func NewUstrFacade() *UstrFacade {
	return &UstrFacade{dummy: 37}
}

func (uf *UstrFacade) Alloc(s string) Ustr {
	return StringToUstr(s)
}
