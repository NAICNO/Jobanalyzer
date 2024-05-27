// The "add" command adds information to the database.  It reads its input from a provided stream.
// This command is remotable.
//
// Major operations:
//
//  add -sample
//    Add sonar sample data.  The default format is "free csv", ie csv with name=value field syntax
//    and no fixed colums.  There are no alternate formats at this time.
//
//  add -sysinfo
//    Add sonar sysinfo data.  The default format is JSON.  There are no alternate formats at this
//    time.

package add

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"go-utils/config"
	. "sonalyze/command"
	"sonalyze/sonarlog"
)

var addHelp = []string{
	"Add new data to the database.  Data are read from stdin, the type and",
	"format are implied by operations -sample or -sysinfo, one of which must",
	"be specified.",
}

type AddCommand struct /* implements RemotableCommand */ {
	DevArgs
	VerboseArgs
	DataDirArgs
	RemotingArgs
	Sample  bool
	Sysinfo bool
}

func (ac *AddCommand) Summary() []string {
	return addHelp
}

func (ac *AddCommand) Add(fs *flag.FlagSet) {
	ac.DevArgs.Add(fs)
	ac.VerboseArgs.Add(fs)
	ac.DataDirArgs.Add(fs)
	ac.RemotingArgs.Add(fs)
	fs.BoolVar(&ac.Sample, "sample", false, "Insert sonar sample data from stdin (zero or more records)")
	fs.BoolVar(&ac.Sysinfo, "sysinfo", false, "Insert sonar sysinfo data from stdin (exactly one record)")
}

func (ac *AddCommand) Validate() error {
	e1 := ac.DevArgs.Validate()
	e4 := ac.VerboseArgs.Validate()
	e2 := ac.RemotingArgs.Validate()
	var e3 error
	if ac.Remoting {
		if ac.DataDir != "" {
			e3 = errors.New("-data-dir may not be used with -remote or -cluster")
		}
	} else {
		e3 = ac.DataDirArgs.Validate()
		if ac.DataDir == "" {
			e3 = errors.Join(e3, errors.New("Required -data-dir argument is absent"))
		}
	}
	var e5 error
	if ac.Sample == ac.Sysinfo {
		e5 = errors.New("One of -sample or -sysinfo must be requested, and not both")
	}
	return errors.Join(e1, e2, e3, e4, e5)
}

func (ac *AddCommand) ReifyForRemote(x *Reifier) error {
	e1 := ac.DevArgs.ReifyForRemote(x)
	// VerboseArgs, DataDirArgs, and RemotingArgs aren't used in remoting and all error checking has
	// already been performed.
	x.String("cluster", ac.Cluster)
	x.Bool("sample", ac.Sample)
	x.Bool("sysinfo", ac.Sysinfo)
	return e1
}

func (ac *AddCommand) AddData(stdin io.Reader, _, _ io.Writer) error {
	data, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	switch {
	case ac.Sample:
		return ac.addSonarFreeCsv(data)
	case ac.Sysinfo:
		return ac.addSysinfo(data)
	default:
		panic("Unexpected")
	}
}

func (ac *AddCommand) addSysinfo(payload []byte) error {
	if ac.Verbose {
		log.Printf("Sysinfo record %d bytes", len(payload))
	}
	var info config.NodeConfigRecord
	err := json.Unmarshal(payload, &info)
	if err != nil {
		return fmt.Errorf("Can't unmarshal Sysinfo JSON: %v", err)
	}
	if info.Timestamp == "" || info.Hostname == "" {
		// Older versions of `sysinfo`
		// TODO: IMPROVEME: Benign if timestamp missing?
		return errors.New("Missing timestamp or host in Sonar sysinfo data")
	}
	ds, err := sonarlog.OpenDir(ac.DataDir)
	if err != nil {
		return err
	}
	defer func() {
		err := ds.Flush()
		if err != nil {
			log.Printf("Flushing data store failed: %v", err)
		}
	}()
	err = ds.AppendBytes(info.Hostname, info.Timestamp, "sysinfo-%s.json", payload)
	if err == sonarlog.BadTimestampErr {
		return nil
	}
	return err
}

func (ac *AddCommand) addSonarFreeCsv(payload []byte) error {
	if ac.Verbose {
		log.Printf("Sample records %d bytes", len(payload))
	}
	ds, err := sonarlog.OpenDir(ac.DataDir)
	if err != nil {
		return err
	}
	defer func() {
		err := ds.Flush()
		if err != nil {
			log.Printf("Flushing data store failed: %v", err)
		}
	}()
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	var result error
	for scanner.Scan() {
		count++
		text := scanner.Text()
		fields, err := getCsvFields(text)
		if err != nil {
			return fmt.Errorf("Can't unmarshal Sonar free CSV: %v", err)
		}
		host := fields["host"]
		time := fields["time"]
		if host == "" || time == "" {
			// TODO: IMPROVEME: Benign if timestamp missing?
			return errors.New("Missing timestamp or host in Sonar sample data")
		}
		err = ds.AppendString(host, time, "%s.csv", text)
		if err != nil && err != sonarlog.BadTimestampErr {
			result = errors.Join(result, err)
		}
	}
	if ac.Verbose {
		log.Printf("Sample records: %d", count)
	}
	return result
}

// Given one line of text on free csv format, return the pairs of field names and values.
//
// Errors:
// - If the CSV reader returns an error err, returns (nil, err), including io.EOF.
// - If any field is seen not to have a field name, return (fields, errNoName) with
//   fields that were valid.

func getCsvFields(text string) (map[string]string, error) {
	rdr := csv.NewReader(strings.NewReader(text))
	rdr.FieldsPerRecord = -1 // Free form, though should not matter
	fields, err := rdr.Read()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, f := range fields {
		ix := strings.IndexByte(f, '=')
		if ix == -1 {
			err = errNoName
			continue
		}
		// TODO: I guess we should detect duplicates?
		result[f[0:ix]] = f[ix+1:]
	}
	return result, err
}

var (
	errNoName = errors.New("CSV field without a field name")
)
