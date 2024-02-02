// Extract system information for sonalyze

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"go-utils/process"
	"go-utils/sysinfo"
)

const (
	KiB = 1024
	MiB = 1024 * 1024
	GiB = 1024 * 1024 * 1024
)

func main() {
	var isNvidia, isAmd bool
	flag.BoolVar(&isNvidia, "nvidia", false, "Get info for Nvidia GPUs")
	flag.BoolVar(&isAmd, "amd", false, "Get info for AMD (ROCm) GPUs")
	flag.Parse()

	model, sockets, coresPerSocket, threadsPerCore, err := sysinfo.CpuInfo()
	if err != nil {
		panic(err)
	}
	memBy, err := sysinfo.PhysicalMemoryBy()
	if err != nil {
		panic(err)
	}
	gpuModel := ""
	gpuCards := 0
	gpuMemBy := int64(0)
	switch {
	case isNvidia:
		gpuModel, gpuCards, gpuMemBy = nvidiaInfo()
	case isAmd:
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
	r.Hostname, err = os.Hostname()
	if err != nil {
		panic("Hostname")
	}
	ht := ""
	if threadsPerCore > 1 {
		ht = " (hyperthreaded)"
	}
	r.MemGB = int64(math.Round(float64(memBy) / GiB))
	r.Model = fmt.Sprintf("%dx%d%s %s, %d GB", sockets, coresPerSocket, ht, model, r.MemGB)
	r.Cores = sockets * coresPerSocket * threadsPerCore
	if gpuModel != "" {
		r.Model += fmt.Sprintf(", %dx %s @ %dGB", gpuCards, gpuModel, gpuMemBy/GiB)
		r.GpuCards = gpuCards
		r.GpumemGB = int64(math.Round((float64(gpuMemBy) * float64(gpuCards)) / GiB))
	}
	bytes, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		panic("Marshal")
	}
	fmt.Println(string(bytes))
}

func nvidiaInfo() (modelName string, cards int, memPerCardBy int64) {
	outside := true
	for _, l := range run("nvidia-smi", "-a") {
		l = strings.TrimSpace(l)
		if outside && strings.HasPrefix(l, "Product Name") {
			_, after, _ := strings.Cut(l, ":")
			modelName = strings.TrimSpace(after)
			cards++
			continue
		}
		if outside && strings.HasPrefix(l, "FB Memory Usage") {
			outside = false
			continue
		}
		if !outside && strings.HasPrefix(l, "Total") {
			_, after, _ := strings.Cut(l, ":")
			after = strings.TrimSpace(strings.TrimSuffix(after, "MiB"))
			n, err := strconv.ParseInt(after, 10, 64)
			if err != nil {
				panic(err)
			}
			memPerCardBy = n * MiB
		}
		outside = true
	}

	return
}

func run(command string, arguments ...string) []string {
	stdout, _, err := process.RunSubprocess(command, arguments)
	if err != nil {
		return []string{}
	}
	return strings.Split(stdout, "\n")
}
