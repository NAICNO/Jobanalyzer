package sonarlog

import (
	"encoding/json"
	"io"

	"go-utils/config"
)

// Sysinfo records appear in sequence in the input without preamble/postamble or separators.

func ParseSysinfoLog(input io.Reader, verbose bool) (records []*config.NodeConfigRecord, err error) {
	records = make([]*config.NodeConfigRecord, 0)
	dec := json.NewDecoder(input)

	for dec.More() {
		var r config.NodeConfigRecord
		err = dec.Decode(&r)
		if err != nil {
			return
		}
		records = append(records, &r)
	}

	return
}
