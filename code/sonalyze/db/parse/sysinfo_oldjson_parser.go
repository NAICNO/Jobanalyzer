// Parser for JSON files holding Sonar `sysinfo` data.

package parse

import (
	"encoding/json"
	"io"

	"go-utils/config"
	"sonalyze/db/repr"
)

// Sysinfo records appear in sequence in the input without preamble/postamble or separators.
//
// If an error is encountered we will return the records successfully parsed along with an error,
// but there is no ability to skip erroneous records and continue going after an error has been
// encountered.

func ParseSysinfoOldJSON(input io.Reader, verbose bool) (records []*repr.SysinfoData, err error) {
	records = make([]*repr.SysinfoData, 0)
	dec := json.NewDecoder(input)

	for dec.More() {
		var r config.NodeConfigRecord
		err = dec.Decode(&r)
		if err != nil {
			return
		}
		records = append(records, (*repr.SysinfoData)(&r))
	}

	return
}
