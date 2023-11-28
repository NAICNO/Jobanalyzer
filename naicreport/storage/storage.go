// Storage management for naicreport.  The file format is "free CSV" form, that is, files use CSV
// syntax but each row can have a different number of columns and each column value starts with
// `<fieldname>=`, so column order is irrelevant.
//
// I/O errors are propagated to the caller.
//
// Rows that appear to be illegal on input are silently dropped.

package storage

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"go-utils/sonarlog"
)

// Given the (relative) name of a root directory, a start date, a date past the end date, and a glob
// pattern, find and return all files that match the pattern in the data store, filtering by the
// start date.  The returned names are relative to the data_path.
//
// The path shall be a clean, absolute path that ends in `/` only if the entire path is `/`.
//
// For the dates, only year/month/day are considered, and timestamps should be passed as UTC times
// with hour, minute, second, and nsec as zero.
//
// The pattern shall have no path components and is typically a glob

func EnumerateFiles(data_path string, from time.Time, to time.Time, pattern string) ([]string, error) {
	filesys := os.DirFS(data_path)
	result := []string{}
	for from.Before(to) {
		probe_fn := fmt.Sprintf("%4d/%02d/%02d/%s", from.Year(), from.Month(), from.Day(), pattern)
		matches, err := fs.Glob(filesys, probe_fn)
		if err != nil {
			return nil, err
		}
		result = append(result, matches...)
		from = from.AddDate(0, 0, 1)
	}
	return result, nil
}

// General "free CSV" reader, returns array of maps from field names to field values.
//
// If the file can't be opened the error with be of type os.PathError.  If there is a parse error
// then the error will be of type encoding.csv.ParseError.  Otherwise the error will be something
// else, most likely an I/O error.

func ReadFreeCSV(filename string) ([]map[string]string, error) {
	input_file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	input := bufio.NewReader(input_file)
	rows, err := ParseFreeCSV(input)
	if err != nil {
		return nil, err
	}
	input_file.Close()
	return rows, nil
}

// This will propagate any errors from the reader; if the reader can't error out (other than EOF),
// then no errors will be returned.

func ParseFreeCSV(input io.Reader) ([]map[string]string, error) {
	rdr := csv.NewReader(input)
	// Rows arbitrarily wide, and possibly uneven.
	rdr.FieldsPerRecord = -1
	rows := make([]map[string]string, 0)
	for {
		fields, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		m := make(map[string]string)
		for _, f := range fields {
			ix := strings.IndexByte(f, '=')
			if ix == -1 {
				// Illegal syntax, just drop the field.
				continue
			}
			m[f[:ix]] = f[ix+1:]
		}
		rows = append(rows, m)
	}
	return rows, nil
}

// General "free CSV" writer.  The fields that are named by `fields` will be written, if they exist
// in the map (otherwise nothing is written for the field).  The fields are written in the order
// given.

func WriteFreeCSV(filename string, fields []string, data []map[string]string) error {
	output_file, err := os.CreateTemp(path.Dir(filename), "naicreport-csvdata")
	if err != nil {
		return err
	}
	wr := csv.NewWriter(output_file)
	for _, row := range data {
		// TODO: With go 1.21, we can hoist this and clear() it after the write, instead of
		// reallocating each time through the loop.
		r := []string{}
		for _, field_name := range fields {
			if field_value, present := row[field_name]; present {
				r = append(r, field_name+"="+field_value)
			}
		}
		if len(r) > 0 {
			wr.Write(r)
		}
	}
	wr.Flush()
	oldname := output_file.Name()
	output_file.Close()
	os.Rename(oldname, filename)
	return nil
}

// The field getters take a string->string map and return the parsed field value of the appropriate
// type (or a compatible zero value), setting *success to false if the field could not be gotten or
// parsed.

// String field.

func GetString(record map[string]string, tag string, success *bool) string {
	value, found := record[tag]
	*success = *success && found
	return value
}

// Job+mark field: a job# optionally suffixed by '<', '>', or '!'.  Drop the suffix.

func GetJobMark(record map[string]string, tag string, success *bool) uint32 {
	s, found := record[tag]
	*success = *success && found
	value, err := strconv.ParseUint(strings.TrimRight(s, "<>!"), 10, 32)
	*success = *success && err == nil
	return uint32(value)
}

// Uint32 field

func GetUint32(record map[string]string, tag string, success *bool) uint32 {
	s, found := record[tag]
	*success = *success && found
	value, err := strconv.ParseUint(s, 10, 32)
	*success = *success && err == nil
	return uint32(value)
}

// Float64 field

func GetFloat64(record map[string]string, tag string, success *bool) float64 {
	s, found := record[tag]
	*success = *success && found
	value, err := strconv.ParseFloat(s, 64)
	*success = *success && err == nil
	return value
}

// Bool field

func GetBool(record map[string]string, tag string, success *bool) bool {
	s, found := record[tag]
	*success = *success && found
	value, err := strconv.ParseBool(s)
	*success = *success && err == nil
	return value
}

// DateTime field.  The logs use this format uniformly (for better or worse).

func GetDateTime(record map[string]string, tag string, success *bool) time.Time {
	s, found := record[tag]
	*success = *success && found
	value, err := time.Parse(sonarlog.DateTimeFormat, s)
	*success = *success && err == nil
	return value
}

// Time field on RFC3339 format

func GetRFC3339(record map[string]string, tag string, success *bool) time.Time {
	s, found := record[tag]
	*success = *success && found
	value, err := time.Parse(time.RFC3339, s)
	*success = *success && err == nil
	return value
}
