// Extract system information for sonar/sonalyze/naicreport/et al.

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-utils/config"
	"go-utils/process"
	"go-utils/sysinfo"
)

const (
	KiB = 1024
	MiB = 1024 * 1024
	GiB = 1024 * 1024 * 1024
)

type gpuCard struct {
	model string
	memBy int64
}

type gpuCards []gpuCard

func (g gpuCards) Len() int { return len(g) }
func (g gpuCards) Less(i, j int) bool {
	if g[i].model == g[j].model {
		return g[i].memBy < g[j].memBy
	}
	return g[i].model < g[j].model
}
func (g gpuCards) Swap(i, j int) { g[i], g[j] = g[j], g[i] }

func main() {
	model, sockets, coresPerSocket, threadsPerCore, err := sysinfo.CpuInfo()
	if err != nil {
		panic(err)
	}

	memBy, err := sysinfo.PhysicalMemoryBy()
	if err != nil {
		panic(err)
	}

	var cards gpuCards
	if nvidiaPath, err := exec.LookPath("nvidia-smi"); err == nil {
		cards = nvidiaInfo(nvidiaPath)
	} else if amdPath, err := exec.LookPath("rocm-smi"); err == nil {
		cards = amdInfo(amdPath)
	}

	var r config.NodeConfigRecord
	r.Timestamp = time.Now().Format(time.RFC3339)

	r.Hostname, err = os.Hostname()
	if err != nil {
		panic("Hostname")
	}

	ht := ""
	if threadsPerCore > 1 {
		ht = " (hyperthreaded)"
	}
	r.MemGB = int(math.Round(float64(memBy) / GiB))
	r.Description = fmt.Sprintf("%dx%d%s %s, %d GB", sockets, coresPerSocket, ht, model, r.MemGB)
	r.CpuCores = sockets * coresPerSocket * threadsPerCore
	if len(cards) > 0 {
		// Bucket the cards that are the same in the description
		sort.Sort(cards)
		i := 0
		for i < len(cards) {
			first := i
			for i++; i < len(cards) && cards[i] == cards[first]; i++ {
			}
			memsize := ""
			if cards[first].memBy > 0 {
				memsize = fmt.Sprint(cards[first].memBy / GiB)
			} else {
				memsize = "unknown "
			}
			r.Description += fmt.Sprintf(", %dx %s @ %sGB", (i - first), cards[first].model, memsize)
		}
		r.GpuCards = len(cards)
		totalMemBy := int64(0)
		for i := 0; i < len(cards); i++ {
			totalMemBy += cards[i].memBy
		}
		r.GpuMemGB = int(totalMemBy / GiB)
	}
	bytes, err := json.MarshalIndent(r, "", " ")
	if err != nil {
		panic("Marshal")
	}
	fmt.Println(string(bytes))
}

// `nvidia-smi -a` dumps a lot of information about all the cards in a semi-structured form,
// each line a textual keyword/value pair.
//
// "Product Name" names the card.  Following the string "FB Memory Usage", "Total" has the
// memory of the card.
//
// Parsing all the output lines in order yield the information about all the cards.

func nvidiaInfo(nvidiaSmiPath string) gpuCards {
	var (
		productNameRe   = regexp.MustCompile(`^\s*Product Name\s*:\s*(.*)$`)
		fbMemoryUsageRe = regexp.MustCompile(`^\s*FB Memory Usage\s*$`)
		totalRe         = regexp.MustCompile(`^\s*Total\s*:\s*(\d+)\s*MiB\s*$`)
		cards           = make(gpuCards, 0)
		lookingForTotal = false
		modelName       = ""
	)
	for _, l := range run(nvidiaSmiPath, "-a") {
		bs := []byte(l)
		if lookingForTotal {
			if ms := totalRe.FindSubmatch(bs); ms != nil {
				n, err := strconv.ParseInt(string(ms[1]), 10, 64)
				if err != nil {
					panic(err)
				}
				if modelName != "" {
					cards = append(cards, gpuCard{modelName, n * MiB})
					modelName = ""
				}
				continue
			}
		} else {
			if ms := productNameRe.FindSubmatch(bs); ms != nil {
				modelName = string(ms[1])
				continue
			}
			if fbMemoryUsageRe.Match(bs) {
				lookingForTotal = true
				continue
			}
		}
		lookingForTotal = false
	}
	return cards
}

// We only have one machine with AMD GPUs at UiO and rocm-smi is unable to show eg how much memory
// is installed on each card on this machine, so this is pretty limited.  But we are at least able
// to extract gross information about the installed cards.
//
// `rocm-smi --showproductname` lists the cards.  The "Card series" line has the card number and
// model name.  There is no memory information, so record it as zero.
//
// TODO: It may be possible to find memory sizes using lspci.  Run `lspci -v` and capture the
// output.  Now look for the line "Kernel modules: amdgpu".  The lines that are part of that block
// of info will have a couple of `Memory at ... ` lines that have memory block sizes, and the first
// line of the info block will have the GPU model.  The largest memory block size is likely the one
// we want.
//
// (It does not appear that the lspci trick works with the nvidia cards - the memory block sizes are
// too small.  This is presumably all driver dependent.)

func amdInfo(rocmSmiPath string) gpuCards {
	var (
		cardSeriesRe = regexp.MustCompile(`^GPU\[(\d+)\].*Card series:\s*(.*)$`)
		cards        = make(gpuCards, 0)
	)
	for _, l := range run(rocmSmiPath, "--showproductname") {
		if ms := cardSeriesRe.FindSubmatch([]byte(l)); ms != nil {
			cards = append(cards, gpuCard{string(ms[2]), 0})
		}
	}
	return cards
}

func run(command string, arguments ...string) []string {
	stdout, _, err := process.RunSubprocess(command, arguments)
	if err != nil {
		return []string{}
	}
	return strings.Split(stdout, "\n")
}
