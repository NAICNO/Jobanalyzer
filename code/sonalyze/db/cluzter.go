// Data representation of Cluzter data, and parser for CSV files holding those data.
//
// The name "Cluzter" is used for the sinfo-derived cluster data since the name "cluster" was
// already used extensively by the database layer for other things.

package db

import (
	"github.com/NordicHPC/sonar/util/formats/newfmt"
)

type CluzterInfo = newfmt.ClusterAttributes

