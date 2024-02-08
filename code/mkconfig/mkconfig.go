// This creates a system config file for a cluster based on raw data collected by sysinfo.
//
// sysinfo can monitor a system ...
//
// Typical usage:
//
// mkconfig \
//    -data-dir ~/sonar/data/mlx.hpc.uio.no \
//    -from 2023-01-01 \
//    -aux ~/.../misc/mlx.hpc.uio.no/mlx.hpc.uio.no-background.json
//

// The format of a sysinfo-HOSTNAME.json file is a sequence of JSON objects without separators.
// (Though most files will have a single object).


package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"go-utils/config"
)

type data struct {
	Hi int `json:"hi"`
	Ho int `json:"ho"`
}

// Command-line parameters
var (
	dataDir string
	from time.Time
	to time.Time
)

func main() {
	test := `{"hi":10}{"ho":20}`
	dec := json.NewDecoder(strings.NewReader(test))
	for {
		var d data
		err := dec.Decode(&d)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(d)
	}
}
