package add

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"go-utils/config"
	. "sonalyze/cmd"
	. "sonalyze/common"
	"sonalyze/db"
	"sonalyze/sonarlog"
)

type AddCommand struct /* implements RemotableCommand */ {
	DevArgs
	VerboseArgs
	DataDirArgs
	RemotingArgs
	ConfigFileArgs
	Sample     bool
	Sysinfo    bool
	SlurmSacct bool
}

//go:embed summary.txt
var summary string

func (ac *AddCommand) Summary(out io.Writer) {
	fmt.Fprint(out, summary)
}

func (ac *AddCommand) Add(fs *CLI) {
	ac.DevArgs.Add(fs)
	ac.VerboseArgs.Add(fs)
	ac.DataDirArgs.Add(fs)
	ac.RemotingArgs.Add(fs)
	ac.ConfigFileArgs.Add(fs)

	fs.Group("data-target")
	fs.BoolVar(&ac.Sample, "sample", false,
		"Insert sonar sample data from stdin (zero or more records)")
	fs.BoolVar(&ac.Sysinfo, "sysinfo", false,
		"Insert sonar sysinfo data from stdin (exactly one record)")
	fs.BoolVar(&ac.SlurmSacct, "slurm-sacct", false,
		"Insert sacct data from stdin (zero or more records)")
}

func (ac *AddCommand) Validate() error {
	var e1, e2, e3, e4, e5, e6 error
	e1 = ac.DevArgs.Validate()
	e4 = ac.VerboseArgs.Validate()
	e2 = ac.RemotingArgs.Validate()
	e6 = ac.ConfigFileArgs.Validate()
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
	opts := 0
	if ac.Sample {
		opts++
	}
	if ac.Sysinfo {
		opts++
	}
	if ac.SlurmSacct {
		opts++
	}
	if opts != 1 {
		e5 = errors.New("Exactly one of -sample, -sysinfo, -slurm-sacct must be requested")
	}
	return errors.Join(e1, e2, e3, e4, e5, e6)
}

func (ac *AddCommand) ReifyForRemote(x *ArgReifier) error {
	e1 := ac.DevArgs.ReifyForRemote(x)
	// VerboseArgs, DataDirArgs, and RemotingArgs aren't used in remoting and all error checking has
	// already been performed.
	x.String("cluster", ac.Cluster)
	x.Bool("sample", ac.Sample)
	x.Bool("sysinfo", ac.Sysinfo)
	x.Bool("slurm-sacct", ac.SlurmSacct)
	return e1
}

func (ac *AddCommand) Perform(stdin io.Reader, _, _ io.Writer) error {
	data, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	switch {
	case ac.Sample:
		return ac.addSonarFreeCsv(data)
	case ac.Sysinfo:
		return ac.addSysinfo(data)
	case ac.SlurmSacct:
		return ac.addSlurmSacctFreeCsv(data)
	default:
		panic("Unexpected")
	}
}

func (ac *AddCommand) addSysinfo(payload []byte) error {
	if ac.Verbose {
		Log.Infof("Sysinfo record %d bytes", len(payload))
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
	cfg, err := db.MaybeGetConfig(ac.ConfigFile())
	if err != nil {
		return err
	}
	ds, err := db.OpenPersistentCluster(ac.DataDir, cfg)
	if err != nil {
		return err
	}
	defer ds.FlushAsync()
	err = ds.AppendSysinfoAsync(info.Hostname, info.Timestamp, payload)
	if err == sonarlog.BadTimestampErr {
		return nil
	}
	return err
}

func (ac *AddCommand) addSonarFreeCsv(payload []byte) error {
	if ac.Verbose {
		Log.Infof("Sample records %d bytes", len(payload))
	}
	cfg, err := db.MaybeGetConfig(ac.ConfigFile())
	if err != nil {
		return err
	}
	ds, err := db.OpenPersistentCluster(ac.DataDir, cfg)
	if err != nil {
		return err
	}
	defer ds.FlushAsync()
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
			// TODO: IMPROVEME: Benign if timestamp missing (would have to drop data)?
			return errors.New("Missing timestamp or host in Sonar sample data")
		}
		err = ds.AppendSamplesAsync(host, time, text)
		if err != nil && err != sonarlog.BadTimestampErr {
			result = errors.Join(result, err)
		}
	}
	if ac.Verbose {
		Log.Infof("Sample records: %d", count)
	}
	return result
}

func (ac *AddCommand) addSlurmSacctFreeCsv(payload []byte) error {
	if ac.Verbose {
		Log.Infof("Sacct records %d bytes", len(payload))
	}
	cfg, err := db.MaybeGetConfig(ac.ConfigFile())
	if err != nil {
		return err
	}
	ds, err := db.OpenPersistentCluster(ac.DataDir, cfg)
	if err != nil {
		return err
	}
	defer ds.FlushAsync()
	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(payload))
	var result error
	for scanner.Scan() {
		count++
		text := scanner.Text()
		fields, err := getCsvFields(text)
		if err != nil {
			return fmt.Errorf("Can't unmarshal sacct free CSV: %v", err)
		}
		// Data are stored in the time-based database according to the End time, which we expect
		// always to have because we only look at completed jobs.
		time := fields["End"]
		if time == "" {
			// TODO: IMPROVEME: Benign if timestamp missing (would have to drop data)?
			return errors.New("Missing timestamp in sacct data")
		}
		err = ds.AppendSlurmSacctAsync(time, text)
		if err != nil && err != sonarlog.BadTimestampErr {
			result = errors.Join(result, err)
		}
	}
	if ac.Verbose {
		Log.Infof("Sacct records: %d", count)
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
	// MT: Constant after initialization; immutable
	errNoName = errors.New("CSV field without a field name")
)
