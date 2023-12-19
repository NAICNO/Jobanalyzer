// The alias file maps a short name to a full name, eg, "ml" => "mlx.hpc.uio.no", and should be used
// throughout the system to allow short names to be used.  There can be multiple short names mapping
// to the same long name.
//
// The input is a json file with an array of mappings:
//  [{"alias":"ml", "value":"mlx.hpc.uio.no"}, ...]

package alias

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Aliases struct {
	filepath string
	mapping map[string]string
}

func ReadAliases(filepath string) (*Aliases, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("No accessible alias file: %v\n", filepath)
	}
	defer f.Close()
	all, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Failed to read alias file: %v\n", filepath)
	}

	type aliasEncoding struct {
		Alias string `json:"alias"`
		Value string `json:"value"`
	}

	var aliases []aliasEncoding
	err = json.Unmarshal(all, &aliases)
	if err != nil {
		return nil, errors.New("Failed to unmarshal aliases")
	}

	mapping := make(map[string]string)
	for _, e := range aliases {
		mapping[e.Alias] = e.Value
	}
	return &Aliases{
		filepath,
		mapping,
	}, nil
}

func (a *Aliases) Resolve(alias string) string {
	// TODO: Requires a mutex or atomic when Reread is implemented
	if m, found := a.mapping[alias]; found {
		return m
	}
	return alias
}

func (a *Aliases) Reread() error {
	// TODO: Implement this, under a mutex or atomic
	return errors.New("alias.Reread not implemented")
}
