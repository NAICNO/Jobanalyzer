package sysinfo

import (
	"fmt"
	"strconv"
	"strings"
	"go-utils/filesys"
)

func CpuInfo() (modelName string, sockets, coresPerSocket, threadsPerCore int) {
	physids := make(map[int]bool)
	siblings := 0
	lines, err := filesys.FileLines("/proc/cpuinfo")
	if err != nil {
		panic(err.Error())
	}
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "model name"):
			modelName = TextField(l)
		case strings.HasPrefix(l, "physical id"):
			physids[int(NumField(l))] = true
		case strings.HasPrefix(l, "siblings"):
			siblings = int(NumField(l))
		case strings.HasPrefix(l, "cpu cores"):
			coresPerSocket = int(NumField(l))
		}
	}
	sockets = len(physids)

	if modelName == "" || sockets == 0 || siblings == 0 || coresPerSocket == 0 {
		panic("Incomplete information in /proc/cpuinfo")
	}

	threadsPerCore = siblings / coresPerSocket
	return
}

func MemInfo() (memSize int64) {
	lines, err := filesys.FileLines("/proc/meminfo")
	if err != nil {
		panic(err.Error())
	}
	for _, l := range lines {
		if strings.HasPrefix(l, "MemTotal:") {
			memSize = NumField(strings.TrimSuffix(l, "kB")) * 1024
			return
		}
	}
	panic("No MemTotal field in /proc/meminfo")
}

// Line must be <whatever>: <text>, return <text> with spaces trimmed
func TextField(s string) string {
	if _, after, found := strings.Cut(s, ":"); found {
		return strings.TrimSpace(after)
	}
	panic(fmt.Sprintf("Bad line: %s", s))
}

// Line must be <whatever>: <text>, return <text> converted to int64
func NumField(s string) int64 {
	if _, after, found := strings.Cut(s, ":"); found {
		x, err := strconv.ParseInt(strings.TrimSpace(after), 10, 64)
		if err == nil {
			return x
		}
	}
	panic(fmt.Sprintf("Bad line: %s", s))
}

