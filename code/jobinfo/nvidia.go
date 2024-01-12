// Run nvidia-smi and return a vector of process samples.
//
// The information is keyed by (device, pid) so that if a process uses multiple devices, the total
// utilization for the process must be summed across devices.  (This is the natural mode of output
// for `nvidia-smi pmon`.)
//
// Crucially, the data are sampling data: they contain no (long) running averages, but are
// snapshots of the system at the time the sample is taken.

package main

import (
	"errors"
	"strconv"
	"strings"
)

const (
	timeoutSeconds = 10
)

type userInfo struct {
	uid uint
	name string
}

// FIXME: these two

var couldNotStart = errors.New("Could not start program")

func safeCommand(command []string, timeout int64) ([]byte, error) {
	panic("NYI")
}

// error really means the command started running but failed, for the reason given.  If the
// command could not be found, we return an empty vector.

func getNvidiaInformation(userByPid map[uint]userInfo) ([]*gpuInfo, error) {
	pmonOut, err := safeCommand(nvidiaPmonCommand, timeoutSeconds)
	if err != nil {
		if err == couldNotStart {
			return make([]*gpuInfo, 0), nil
		}
		return nil, err
	}
	processes, err := parseNvidiaPmonOutput(pmonOut, userByPid)
	if err != nil {
		return nil, err
	}
	queryOut, err := safeCommand(nvidiaQueryCommand, timeoutSeconds)
	if err != nil {
		return nil, err
	}
	qprocs, err := parseNvidiaQueryOutput(queryOut, userByPid)
	if err != nil {
		return nil, err
	}
	processes = append(processes, qprocs...)
	return processes, nil
}

// For prototyping purposes (and maybe it's good enough for production?), parse the output of
// `nvidia-smi pmon`.  This output has a couple of problems:
//
//  - it is (documented to be) not necessarily stable
//  - it does not orphaned processes holding onto GPU memory, the way nvtop can do
//
// To fix the latter problem we do something with --query-compute-apps, see later.
//
// Note that `-c 1 -s u` gives us more or less instantaneous utilization, not some long-running
// average.
//
// TODO: We could consider using the underlying C library instead, but this adds a fair
// amount of complexity.  See the nvidia-smi manual page.

var nvidiaPmonCommand = []string{"nvidia-smi", "pmon", "-c", "1", "-s", "mu"}

func parseNvidiaPmonOutput(rawOutput []byte, userByPid map[uint]userInfo) (result []*gpuInfo, err error) {
	result = make([]*gpuInfo, 0)
	for _, line := range strings.Split(string(rawOutput), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		fs := strings.Fields(line)
		pidStr := fs[1]
		if pidStr == "-" {
			continue
		}
		var device, memSizeMiB, pid uint64
		device, err = strconv.ParseUint(fs[0], 10, 32)
		if err != nil {
			return
		}
		memSizeMiB, err = strconv.ParseUint(fs[3], 10, 64)
		if err != nil {
			memSizeMiB = 0
		}
		var gpuPct, memPct float64
		gpuPct, err = strconv.ParseFloat(fs[4], 64)
		if err != nil {
			gpuPct = 0
		}
		memPct, err = strconv.ParseFloat(fs[5], 64)
		if err != nil {
			memPct = 0
		}
		// For nvidia-smi, we use the first word because the command produces blank-padded
		// output.  We can maybe do better by considering non-empty words.
		command := fs[8]
		pid, err = strconv.ParseUint(pidStr, 10, 32)
		if err != nil {
			return
		}
		var user userInfo
		if uinfo, found := userByPid[uint(pid)]; found {
			user = uinfo
		} else {
			user = userInfo{
				name: "_zombie_" + pidStr,
				uid: invalidUID,
			}
		}
		result = append(result, &gpuInfo{
			device: int(device),
			pid: uint(pid),
			user: user.name,
			uid: user.uid,
			gpuPct: gpuPct,
			memPct: memPct,
			memSizeKiB: memSizeMiB * 1024,
			command: command,
		})
	}
	return
}

// We use this to get information about processes that are not captured by pmon.  It's hacky
// but it works.
//
// Same signature as parseNvidiaPmonOutput(), q.v. but `user` is always "_zombie_PID" and `command`
// is always "_unknown_".  Only pids *not* in userByPid are returned.

var nvidiaQueryCommand = []string{"nvidia-smi", "--query-compute-apps=pid,used_memory", "--format=csv,noheader,nounits"}

func parseNvidiaQueryOutput(rawOutput []byte, userByPid map[uint]userInfo) (result []*gpuInfo, err error) {
	result = make([]*gpuInfo, 0)
	for _, line := range strings.Split(string(rawOutput), "\n") {
		fs := strings.Fields(line)
		pidStr, _ := strings.CutSuffix(fs[0], ",")
		var pid uint64
		pid, err = strconv.ParseUint(pidStr, 10, 32)
		if err != nil {
			return
		}
		if _, found := userByPid[uint(pid)]; found {
			continue
		}
		var memUsage uint64
		memUsage, err = strconv.ParseUint(fs[1], 10, 64)
		if err != nil {
			return
		}
		result = append(result, &gpuInfo{
			device: noGpuDevice,
			pid: uint(pid),
			user: "_zombie_" + pidStr,
			uid: invalidUID,
			gpuPct: 0,
			memPct: 0,
			memSizeKiB: memUsage,
			command: "_unknown_",
		})
	}
	return
}
