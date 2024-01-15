// Extract system information for sonalyze

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"

	"go-utils/process"
	"go-utils/sysinfo"
)

func main() {
	var isNvidia, isAmd bool
	flag.BoolVar(&isNvidia, "nvidia", false, "Get info for Nvidia GPUs")
	flag.BoolVar(&isAmd, "amd", false, "Get info for AMD (ROCm) GPUs")
	flag.Parse()

	model, sockets, coresPerSocket, threadsPerCore := sysinfo.CpuInfo()
	mem := sysinfo.MemInfo()
	gpuModel, gpuCards, gpuMem := "", 0, int64(0)
	switch {
	case isNvidia:
		gpuModel, gpuCards, gpuMem = nvidiaInfo()
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
	var err error
	r.Hostname, err = os.Hostname()
	if err != nil {
		panic("Hostname")
	}
	ht := ""
	if threadsPerCore > 1 {
		ht = " (hyperthreaded)"
	}
	r.MemGB = int64(math.Round(float64(mem) / (1024 * 1024 * 1024)))
	r.Model = fmt.Sprintf("%dx%d%s %s, %d GB", sockets, coresPerSocket, ht, model, r.MemGB)
	r.Cores = sockets * coresPerSocket * threadsPerCore
	if gpuModel != "" {
		r.Model += fmt.Sprintf(", %dx %s @ %dGB", gpuCards, gpuModel, gpuMem/(1024*1024*1024))
		r.GpuCards = gpuCards
		r.GpumemGB = int64(math.Round((float64(gpuMem) * float64(gpuCards)) / (1024 * 1024 * 1024)))
	}
	bytes, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		panic("Marshal")
	}
	fmt.Println(string(bytes))
}

func nvidiaInfo() (modelName string, cards int, memPerCard int64) {
	outside := true
	for _, l := range run("nvidia-smi", "-a") {
		l = strings.TrimSpace(l)
		if outside && strings.HasPrefix(l, "Product Name") {
			modelName = sysinfo.TextField(l)
			cards++
			continue
		}
		if outside && strings.HasPrefix(l, "FB Memory Usage") {
			outside = false
			continue
		}
		if !outside && strings.HasPrefix(l, "Total") {
			memPerCard = sysinfo.NumField(strings.TrimSuffix(l, "MiB")) * 1024 * 1024
		}
		outside = true
	}

	return
}

func run(command string, arguments ...string) []string {
	output, err := process.RunSubprocess(command, arguments)
	if err != nil {
		return []string{}
	}
	return strings.Split(output, "\n")
}
