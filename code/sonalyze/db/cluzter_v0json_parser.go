package db

import (
	"io"

	"github.com/NordicHPC/sonar/util/formats/newfmt"
)

func ParseCluzterV0JSON(
	input io.Reader,
	verbose bool,
) (
	records []*CluzterInfo,
	softErrors int,
	err error,
) {
	records = make([]*CluzterInfo, 0)
	err = newfmt.ConsumeJSONCluster(input, false, func(r *newfmt.ClusterEnvelope) {
		if r.Data != nil {
			// TODO: Not optimal to be pointing into that record probably, but not optimal to be
			// using the original encoding in any case.
			records = append(records, &r.Data.Attributes)
		} else {
			softErrors++
		}
	})
	return
}
