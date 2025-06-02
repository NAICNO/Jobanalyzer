// Parser for JSON files holding Sonar `sysinfo` data.

package parse

import (
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"github.com/NordicHPC/sonar/util/formats/oldfmt"
	"sonalyze/db/repr"
)

// The oldfmt description looks like this:
//
//   2x48 (hyperthreaded) AMD EPYC 7642 48-Core Processor, 1007 GiB, 4x NVIDIA GeForce RTX 3090 @ 24GiB

var (
	descMatcher = regexp.MustCompile(`^(\d+)x(\d+)( \(hyperthreaded\))?(.*?), \d+ GiB`)
	gpuMatcher  = regexp.MustCompile(`, \d+x (.*) @ (\d+)GiB$`)
)

// Sysinfo records appear in sequence in the input without preamble/postamble or separators.
//
// If an error is encountered we will return the records successfully parsed along with an error,
// but there is no ability to skip erroneous records and continue going after an error has been
// encountered.

func ParseSysinfoOldJSON(
	input io.Reader,
	verbose bool,
) (
	nodeData []*repr.SysinfoNodeData,
	cardData []*repr.SysinfoCardData,
	softErrors int,
	err error,
) {
	nodeData = make([]*repr.SysinfoNodeData, 0)
	cardData = make([]*repr.SysinfoCardData, 0)
	dec := json.NewDecoder(input)

	for dec.More() {
		var r oldfmt.SysinfoEnvelope
		err = dec.Decode(&r)
		if err != nil {
			return
		}
		var sockets, coresPerSocket, threadsPerCore uint64
		var cpuModel, architecture string
		if m := descMatcher.FindStringSubmatch(r.Description); m != nil {
			sockets, _ = strconv.ParseUint(m[1], 10, 64)
			coresPerSocket, _ = strconv.ParseUint(m[2], 10, 64)
			var threads uint64 = 1
			if m[3] != "" {
				threads = 2
			}
			threadsPerCore = threads
			cpuModel = strings.TrimSpace(m[4])
		}
		// Architecture names as reported by Sonar: src/realsystem.rs
		if strings.Contains(r.Description, "Intel") || strings.Contains(r.Description, "AMD") {
			architecture = "x86_64"
		} else {
			architecture = "aarch64"
		}
		nodeData = append(nodeData, &repr.SysinfoNodeData{
			Time:           r.Timestamp,
			Node:           r.Hostname,
			Memory:         uint64(r.MemGB) * 1024 * 1024,
			OsName:         "Linux",
			Sockets:        sockets,
			CoresPerSocket: coresPerSocket,
			ThreadsPerCore: threadsPerCore,
			CpuModel:       cpuModel,
			Architecture:   architecture,
			// Unknown, the only one remotely important is Cluster, probably
			//   Cluster
			//   OsRelease
			//   TopoSVG
		})

		var model, manufacturer string
		var gpumem uint64
		if m := gpuMatcher.FindStringSubmatch(r.Description); m != nil {
			model = m[1]
			gpumem, _ = strconv.ParseUint(m[2], 10, 64)
			gpumem *= 1024 * 1024
			switch {
			case strings.Contains(model, "NVIDIA"):
				manufacturer = "NVIDIA"
			case strings.Contains(model, "AMD"):
				manufacturer = "AMD"
			case strings.Contains(model, "Intel"):
				manufacturer = "Intel"
			}
		}
		// Prefer GpuInfo, fall back to synthesizing data from GpuCards
		switch {
		case r.GpuInfo != nil:
			for i, o := range r.GpuInfo {
				cardData = append(cardData, &repr.SysinfoCardData{
					Time: r.Timestamp,
					Node: r.Hostname,
					SysinfoGpuCard: &newfmt.SysinfoGpuCard{
						Index:          uint64(i),
						UUID:           o.UUID,
						Address:        o.BusAddress,
						Manufacturer:   o.Manufacturer,
						Model:          o.Model,
						Architecture:   o.Architecture,
						Driver:         o.Driver,
						Firmware:       o.Firmware,
						Memory:         o.MemKB,
						PowerLimit:     o.PowerLimit,
						MaxPowerLimit:  o.MaxPowerLimit,
						MinPowerLimit:  o.MinPowerLimit,
						MaxCEClock:     o.MaxCEClock,
						MaxMemoryClock: o.MaxMemClock,
					},
				})
			}
		case r.GpuCards > 0:
			for i := range r.GpuCards {
				cardData = append(cardData, &repr.SysinfoCardData{
					Time: r.Timestamp,
					Node: r.Hostname,
					SysinfoGpuCard: &newfmt.SysinfoGpuCard{
						Index:        i,
						Model:        model,
						Manufacturer: manufacturer,
						Memory:       gpumem,
					},
				})
			}
		}
	}
	return
}
