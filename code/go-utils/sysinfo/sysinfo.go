// Misc utilities for getting information about the system.
//
// In addition to these there are standard functions:
//
//  os.Hostname() - gets the host name

package sysinfo

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

// UserMap: Optimize the lookup from uid to user name by caching.

type UserMap struct {
	m map[uint]string
}

func NewUserMap() *UserMap {
	return &UserMap{make(map[uint]string)}
}

func (um *UserMap) LookupUid(uid uint) string {
	if probe, found := um.m[uid]; found {
		return probe
	}
	// Raw getpwnam() call
	user, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		return "_noinfo_"
	}
	name := user.Username
	um.m[uid] = name
	return name
}

// EnumeratePids: Find all pids in /proc/pid

type PidAndUid struct {
	Pid, Uid uint
}

func EnumeratePids() ([]PidAndUid, error) {
	dirents, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}
	result := make([]PidAndUid, 0)
	for _, de := range dirents {
		pid, err := strconv.ParseUint(de.Name(), 10, 64)
		if err != nil {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		// Unix only, and pretty ugly - the structure of the Sys object is not properly documented.
		// https://stackoverflow.com/questions/58179647/getting-uid-and-gid-of-a-file
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}
		uid := uint(stat.Uid)
		result = append(result, PidAndUid{Pid: uint(pid), Uid: uid})
	}
	return result, nil
}

// CpuInfo: get information about installed CPUs

func CpuInfo() (modelName string, sockets, coresPerSocket, threadsPerCore int, err error) {
	physids := make(map[int]bool)
	siblings := 0
	lines, err := fileLines("/proc/cpuinfo")
	if err != nil {
		return
	}
	var n int64
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "model name"):
			modelName, err = textField(l)
		case strings.HasPrefix(l, "physical id"):
			n, err = numField(l)
			physids[int(n)] = true
		case strings.HasPrefix(l, "siblings"):
			n, err = numField(l)
			siblings = int(n)
		case strings.HasPrefix(l, "cpu cores"):
			n, err = numField(l)
			coresPerSocket = int(n)
		}
		if err != nil {
			return
		}
	}
	sockets = len(physids)

	if modelName == "" || sockets == 0 || siblings == 0 || coresPerSocket == 0 {
		err = fmt.Errorf("Incomplete information in /proc/cpuinfo")
		return
	}

	threadsPerCore = siblings / coresPerSocket
	return
}

// PhysicalMemoryBy: get size of physical memory in bytes

func PhysicalMemoryBy() (memSize uint64, err error) {
	lines, err := fileLines("/proc/meminfo")
	if err != nil {
		return
	}
	for _, l := range lines {
		if strings.HasPrefix(l, "MemTotal:") {
			var n int64
			n, err = numField(strings.TrimSuffix(l, "kB"))
			memSize = uint64(n) * 1024
			return
		}
	}
	err = fmt.Errorf("No MemTotal field in /proc/meminfo")
	return
}

// PagesizeBy: get the size of pages in bytes.  This is a little silly.

func PagesizeBy() (pageSize uint) {
	return uint(os.Getpagesize())
}

// BootTime: get the time of boot in seconds since Unix epoch

func BootTime() (btime int64, err error) {
	lines, err := fileLines("/proc/stat")
	if err != nil {
		return
	}
	for _, l := range lines {
		if strings.HasPrefix(l, "btime ") {
			btime, err = strconv.ParseInt(strings.Fields(l)[1], 10, 64)
			return
		}
	}
	err = fmt.Errorf("Could not find btime in /proc/stat")
	return
}

func fileLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not open %s: %v", filename, err)
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Could not read %s: %v", filename, err)
	}
	return strings.Split(string(bytes), "\n"), nil
}

// Line must be <whatever>: <text>, return <text> with spaces trimmed

func textField(s string) (string, error) {
	if _, after, found := strings.Cut(s, ":"); found {
		return strings.TrimSpace(after), nil
	}
	return "", fmt.Errorf("Bad line: %s", s)
}

// Line must be <whatever>: <text>, return <text> converted to int64

func numField(s string) (int64, error) {
	if _, after, found := strings.Cut(s, ":"); found {
		return strconv.ParseInt(strings.TrimSpace(after), 10, 64)
	}
	return 0, fmt.Errorf("Bad line: %s", s)
}
