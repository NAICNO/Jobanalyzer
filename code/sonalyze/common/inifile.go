package common

import (
	"errors"
	"os"
	"path"

	ini "github.com/lars-t-hansen/ini"
)

// MT: Constant after initialization
var (
	p = ini.NewParser()
	store *ini.Store
	dataSource = p.AddSection("data-source")
	DataSourceRemote = dataSource.AddString("remote")
	DataSourceAuthFile = dataSource.AddString("auth-file")
	DataSourceCluster = dataSource.AddString("cluster")
	DataSourceDataDir = dataSource.AddString("data-dir")
	DataSourceFrom = dataSource.AddString("from")
	DataSourceTo = dataSource.AddString("to")
)

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		return
	}
	fn := path.Join(path.Clean(home), ".sonalyze")
	input, err := os.Open(fn)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			Log.Errorf("Error in trying to open %s: %s", fn, err.Error())
		}
		return
	}
	defer input.Close()
	store, err = p.Parse(input)
	if err != nil {
		Log.Errorf("Error in trying to parse %s: %s", fn, err.Error())
		return
	}
}

func HasDefault(f *ini.Field) bool {
	return store != nil && f.Present(store)
}

func ApplyDefault(sp *string, f *ini.Field) bool {
	if *sp != "" || store == nil || !f.Present(store) {
		return false
	}
	*sp = os.ExpandEnv(f.StringVal(store))
	return true
}
