// SPDX-License-Identifier: MIT
// From https://github.com/lars-t-hansen/util/go-utils

package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
)

func CsvLines(source any, cb func([]string)) {
	var input io.Reader
	switch x := source.(type) {
	case io.Reader:
		input = x
	case string:
		infile, err := os.Open(x)
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()
		input = infile
	default:
		log.Fatal("Bad type to CsvLines")
	}
	r := csv.NewReader(input)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		cb(record)
	}
}
