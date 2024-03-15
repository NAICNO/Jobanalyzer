// Interned string type, this is called "Ustr" because that's what it's called in the Rust code.
//
// There is a global, thread-safe store of Ustr values.  Casual uses can use StringToUstr() to map a
// string to its Ustr.  For performance-sensitive uses, there are a couple of other options:
//
//  - In multi-threaded situations with a lot of string creation the store can become contended.
//    In this case, use UstrCache (see further down) as a thread-local cache
//
//  - Conversions between string and []byte incur allocations in both directions, in principle.  If
//    this becomes a bottleneck, use BytesToUstr() and the AllocBytes method of the cache to avoid
//    these conversions both at the caller and within this code.
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
	"strings"
	"sync"
)

type Ustr uint32

var (
	tableLock   sync.RWMutex
	internTable hashtable
	revTable    []string
)

// The zero value of Ustr is the empty string

const UstrEmpty Ustr = 0 // StringToUstr("")

func init() {
	internTable = newHashtable()
	revTable = make([]string, 0)
	_ = StringToUstr("")
}

// Return the Ustr for the string.  This is guaranteed not to retain s.

func StringToUstr(s string) Ustr {
	u, _ := stringToUstrAndString(s)
	return u
}

func BytesToUstr(bs []byte) Ustr {
	u, _ := bytesToUstrAndString(bs)
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

	println(internTable.size)
	if printStrings {
		for _, v := range revTable {
			println(v)
		}
	}
}

func stringToUstrAndString(s string) (Ustr, string) {
	h := hashString(s)

	tableLock.RLock()
	if probe := internTable.getString(h, s); probe != nil {
		tableLock.RUnlock()
		return probe.ustr, probe.name
	}
	tableLock.RUnlock()

	tableLock.Lock()
	defer tableLock.Unlock()

	// Maybe it changed while we were unlocked
	if probe := internTable.getString(h, s); probe != nil {
		return probe.ustr, probe.name
	}

	name := strings.Clone(s)
	ustr := Ustr(len(revTable))
	revTable = append(revTable, name)

	internTable.insert(h, name, ustr)
	return ustr, name
}

func bytesToUstrAndString(bs []byte) (Ustr, string) {
	h := hashBytes(bs)

	tableLock.RLock()
	if probe := internTable.getBytes(h, bs); probe != nil {
		tableLock.RUnlock()
		return probe.ustr, probe.name
	}
	tableLock.RUnlock()

	tableLock.Lock()
	defer tableLock.Unlock()

	// Maybe it changed while we were unlocked
	if probe := internTable.getBytes(h, bs); probe != nil {
		return probe.ustr, probe.name
	}

	name := string(bs)
	ustr := Ustr(len(revTable))
	revTable = append(revTable, name)

	internTable.insert(h, name, ustr)
	return ustr, name
}

// An interface for either the caching allocator or the no-op facade.

type UstrAllocator interface {
	Alloc(s string) Ustr
	AllocBytes(bs []byte) Ustr
}

// This is an unsynchronized cache that is a facade for the global Ustr store.  Use this in a
// thread-local way if the global store is very contended.

type UstrCache struct {
	cache hashtable
}

func NewUstrCache() *UstrCache {
	return &UstrCache{cache: newHashtable()}
}

func (uc *UstrCache) Alloc(s string) Ustr {
	h := hashString(s)
	if probe := uc.cache.getString(h, s); probe != nil {
		return probe.ustr
	}
	ustr, name := stringToUstrAndString(s)
	uc.cache.insert(h, name, ustr)
	return ustr
}

func (uc *UstrCache) AllocBytes(bs []byte) Ustr {
	h := hashBytes(bs)
	if probe := uc.cache.getBytes(h, bs); probe != nil {
		return probe.ustr
	}
	ustr, name := bytesToUstrAndString(bs)
	uc.cache.insert(h, name, ustr)
	return ustr
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

func (uf *UstrFacade) AllocBytes(bs []byte) Ustr {
	return BytesToUstr(bs)
}

// hashtable maps a string or []byte to a hashnode, treating the two key types as equivalent.  The
// node carries the name and the Ustr value of that name.

const (
	inverseLoad uint32 = 3
	initialCapacity uint32 = 100
)

type hashtable struct {
	table     []*hashnode
	size      uint32
	divisor   uint32
	remaining uint32
}

type hashnode struct {
	hash hashcode
	name string
	ustr Ustr
	next *hashnode
}

func newHashtable() hashtable {
	size := inverseLoad * initialCapacity
	return hashtable{
		table: make([]*hashnode, size),
		size: 0,
		divisor: size,
		remaining: initialCapacity,
	}
}

func (ht *hashtable) getString(h hashcode, s string) *hashnode {
	slot := uint32(h) % ht.divisor
	probe := ht.table[slot]
	for probe != nil && s != probe.name {
		probe = probe.next
	}
	return probe
}

func (ht *hashtable) getBytes(h hashcode, bs []byte) *hashnode {
	slot := uint32(h) % ht.divisor
	probe := ht.table[slot]
	for probe != nil {
		name := probe.name
		found := true
		if len(name) != len(bs) {
			found = false
		} else {
			for i := range name {
				if name[i] != bs[i] {
					found = false
					break
				}
			}
		}
		if found {
			break
		}
		probe = probe.next
	}
	return probe
}

func (ht *hashtable) insert(h hashcode, name string, ustr Ustr) {
	ht.maybeRehash()
	ht.remaining--
	ht.size++
	slot := uint32(h) % ht.divisor
	node := &hashnode {
		hash: h,
		ustr: ustr,
		name: name,
		next: ht.table[slot],
	}
	ht.table[slot] = node
}

func (ht *hashtable) maybeRehash() {
	if ht.remaining == 0 {
		newSize := 2 * uint32(len(ht.table))
		newRemaining := uint32(len(ht.table)) / inverseLoad
		newTable := make([]*hashnode, newSize)
		for _, l := range ht.table {
			for l != nil {
				p := l
				l = l.next
				slot := uint32(p.hash) % newSize
				p.next = newTable[slot]
				newTable[slot] = p
			}
		}
		ht.table = newTable
		ht.divisor = newSize
		ht.remaining = newRemaining
	}
}

// hashString and hashBytes must return the same values for the same bytes in the same order.

type hashcode uint32

func hashString(s string) hashcode {
	h := uint32(0)
	for i := range s {
		h = (h << 3) ^ uint32(s[i])
	}
	return hashcode(h)
}

func hashBytes(bs []byte) hashcode {
	h := uint32(0)
	for _, c := range bs {
		h = (h << 3) ^ uint32(c)
	}
	return hashcode(h)
}

