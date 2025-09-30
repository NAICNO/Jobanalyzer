// Parser for v0 "new format" JSON files holding Sonar `sysinfo` data.

package parse

import (
	"encoding/base64"
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
			var topoSvg, topoText string
			if d.TopoSVG != "" {
				dst := make([]byte, base64.StdEncoding.DecodedLen(len(d.TopoSVG)))
				n, err := base64.StdEncoding.Decode(dst, []byte(d.TopoSVG))
				if err != nil {
					softErrors++
				}
				topoSvg = string(dst[:n])
			}
			if d.TopoText != "" {
				dst := make([]byte, base64.StdEncoding.DecodedLen(len(d.TopoText)))
				n, err := base64.StdEncoding.Decode(dst, []byte(d.TopoText))
				if err != nil {
					softErrors++
				}
				topoText = string(dst[:n])
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
				TopoSVG:        topoSvg,
				TopoText:       topoText,
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
