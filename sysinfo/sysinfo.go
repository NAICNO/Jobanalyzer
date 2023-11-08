// Extract system information for sonalyze

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	isNvidia := flag.Bool("nvidia", false, "Get info for Nvidia GPUs")
	isAmd := flag.Bool("amd", false, "Get info for AMD (ROCm) GPUs")
	flag.Parse()

	model, sockets, coresPerSocket, threadsPerCore := cpuinfo()
	mem := meminfo()
	gpuModel, gpuCards, gpuMem := "", 0, int64(0)
	switch {
	case *isNvidia:
		gpuModel, gpuCards, gpuMem = nvidiaInfo()
	case *isAmd:
		fmt.Fprintf(os.Stderr, "%s: No AMD support yet\n", os.Args[0])
		os.Exit(1)
	}

	type repr struct {
		Hostname  string `json:"hostname"`
		Model     string `json:"description"`
		Cores     int    `json:"cpu_cores"`
		MemGB     int64  `json:"mem_gb"`
		GpuCards  int    `json:"gpu_cards,omitempty"`
		GpumemGB  int64  `json:"gpumem_gb,omitempty"`
		GpumemPct bool   `json:"gpumem_pct,omitempty"`
	}
	var r repr
	var err error
	r.Hostname, err = os.Hostname()
	if err != nil {
		panic("Hostname")
	}
	ht := ""
	if threadsPerCore > 1 {
		ht = " (hyperthreaded)"
	}
	r.Model = fmt.Sprintf("%dx%d%s %s", sockets, coresPerSocket, ht, model)
	r.Cores = sockets * coresPerSocket * threadsPerCore
	r.MemGB = mem / (1024 * 1024 * 1024)
	if gpuModel != "" {
		r.Model += fmt.Sprintf(", %dx %s @ %dGB", gpuCards, gpuModel, gpuMem/(1024*1024*1024))
		r.GpuCards = gpuCards
		r.GpumemGB = (gpuMem * int64(gpuCards)) / (1024 * 1024 * 1024)
	}
	bytes, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		panic("Marshal")
	}
	fmt.Println(string(bytes))
}

func cpuinfo() (modelName string, sockets, coresPerSocket, threadsPerCore int) {
	physids := make(map[int]bool)
	siblings := 0
	for _, l := range lines("/proc/cpuinfo") {
		switch {
		case strings.HasPrefix(l, "model name"):
			modelName = textField(l)
		case strings.HasPrefix(l, "physical id"):
			physids[int(numField(l))] = true
		case strings.HasPrefix(l, "siblings"):
			siblings = int(numField(l))
		case strings.HasPrefix(l, "cpu cores"):
			coresPerSocket = int(numField(l))
		}
	}

	sockets = len(physids)
	threadsPerCore = siblings / coresPerSocket
	return
}

func meminfo() (memSize int64) {
	for _, l := range lines("/proc/meminfo") {
		if strings.HasPrefix(l, "MemTotal:") {
			memSize = numField(strings.TrimSuffix(l, "kB")) * 1024
			return
		}
	}
	panic("No MemTotal field in /proc/meminfo")
}

func nvidiaInfo() (modelName string, cards int, memPerCard int64) {
	outside := true
	for _, l := range run("nvidia-smi", "-a") {
		l = strings.TrimSpace(l)
		if outside && strings.HasPrefix(l, "Product Name") {
			modelName = textField(l)
			cards++
			continue
		}
		if outside && strings.HasPrefix(l, "FB Memory Usage") {
			outside = false
			continue
		}
		if !outside && strings.HasPrefix(l, "Total") {
			memPerCard = numField(strings.TrimSuffix(l, "MiB")) * 1024 * 1024
		}
		outside = true
	}

	return
}

func textField(s string) string {
	if _, after, found := strings.Cut(s, ":"); found {
		return strings.TrimSpace(after)
	}
	panic(fmt.Sprintf("Bad line: %s", s))
}

func numField(s string) int64 {
	if _, after, found := strings.Cut(s, ":"); found {
		x, err := strconv.ParseInt(strings.TrimSpace(after), 10, 64)
		if err == nil {
			return x
		}
	}
	panic(fmt.Sprintf("Bad line: %s", s))
}

func run(command string, arguments ...string) []string {
	cmd := exec.Command(command, arguments...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	errs := stderr.String()
	if err != nil || errs != "" {
		return []string{}
	}
	return strings.Split(stdout.String(), "\n")
}

func lines(fn string) []string {
	if bytes, err := os.ReadFile(fn); err == nil {
		return strings.Split(string(bytes), "\n")
	}
	panic(fmt.Sprintf("Could not open %s", fn))
}
