// Parser for v0 "new format" JSON files holding Sonar `sysinfo` data.

package parse

import (
	"io"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"sonalyze/db/repr"
)

var defaultDistances = [][]uint64{[]uint64{10}}

func ParseSysinfoV0JSON(
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
	err = newfmt.ConsumeJSONSysinfo(input, false, func(r *newfmt.SysinfoEnvelope) {
		if r.Data != nil {
			var d *newfmt.SysinfoAttributes = &r.Data.Attributes
			distances := d.Distances
			if distances == nil {
				distances = defaultDistances
			}
			nodeData = append(nodeData, &repr.SysinfoNodeData{
				Time:           string(d.Time),
				Cluster:        string(d.Cluster),
				Node:           string(d.Node),
				OsName:         string(d.OsName),
				OsRelease:      string(d.OsRelease),
				Architecture:   string(d.Architecture),
				Sockets:        uint64(d.Sockets),
				CoresPerSocket: uint64(d.CoresPerSocket),
				ThreadsPerCore: uint64(d.ThreadsPerCore),
				CpuModel:       d.CpuModel,
				Memory:         uint64(d.Memory),
				TopoSVG:        d.TopoSVG,
				Distances:      distances,
			})
			for i := range d.Cards {
				cardData = append(cardData, &repr.SysinfoCardData{
					Time:           string(d.Time),
					Node:           string(d.Node),
					SysinfoGpuCard: &d.Cards[i],
				})
			}
		} else {
			softErrors++
		}
	})
	return
}
