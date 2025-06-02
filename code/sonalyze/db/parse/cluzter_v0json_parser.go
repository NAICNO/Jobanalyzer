package parse

import (
	"io"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
	"sonalyze/db/repr"
)

func ParseCluzterV0JSON(
	input io.Reader,
	verbose bool,
) (
	attributes []*repr.CluzterAttributes,
	partitions []*repr.CluzterPartitions,
	nodes []*repr.CluzterNodes,
	softErrors int,
	err error,
) {
	attributes = make([]*repr.CluzterAttributes, 0)
	partitions = make([]*repr.CluzterPartitions, 0)
	nodes = make([]*repr.CluzterNodes, 0)
	err = newfmt.ConsumeJSONCluster(input, false, func(r *newfmt.ClusterEnvelope) {
		if r.Data != nil {
			d := &r.Data.Attributes
			attributes = append(attributes, &repr.CluzterAttributes{
				Time:    string(d.Time),
				Cluster: string(d.Cluster),
				Slurm:   d.Slurm,
			})
			partitions = append(partitions, &repr.CluzterPartitions{
				Time:       string(d.Time),
				Cluster:    string(d.Cluster),
				Partitions: d.Partitions,
			})
			nodes = append(nodes, &repr.CluzterNodes{
				Time:    string(d.Time),
				Cluster: string(d.Cluster),
				Nodes:   d.Nodes,
			})
		} else {
			softErrors++
		}
	})
	return
}
