package common

import (
	"errors"
	"os"
	"path"

	"go-utils/ini"
)

// MT: Constant after initialization
var IniFileMaybe *ini.IniFile

// Attempt to read the ini file.  If it's not there or reading fails, return nil.  Print warnings on
// failures, but never return an error.

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
	ini, err := ini.ParseIni(input)
	if err != nil {
		Log.Errorf("Error in trying to parse %s: %s", fn, err.Error())
		return
	}

	for _, section := range ini {
		for name := range section.Vars {
			// This is probably OK - the spec only says that things get tricky when we insert or
			// remove elements, but here we're updating an existing element.
			section.Vars[name] = os.ExpandEnv(section.Vars[name])
		}
	}

	IniFileMaybe = &ini
}

func HasDefault(section, key string) bool {
	if IniFileMaybe != nil {
		if defaults := (*IniFileMaybe)[section]; defaults != nil {
			if _, found := defaults.Vars[key]; found {
				return true
			}
		}
	}
	return false
}

func ApplyDefault(sp *string, section, key string) bool {
	if IniFileMaybe != nil {
		if *sp == "" {
			if defaults := (*IniFileMaybe)[section]; defaults != nil {
				if val, found := defaults.Vars[key]; found {
					*sp = val
					return true
				}
			}
		}
	}
	return false
}
