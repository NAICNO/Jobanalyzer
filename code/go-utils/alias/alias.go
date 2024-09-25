// Cluster name aliasing abstraction.
//
// The alias file maps a short name to a full name, eg, "ml" => "mlx.hpc.uio.no", and should be used
// throughout the system to allow short names (the aliases) to be used.  There can be multiple short
// names mapping to the same long name.
//
// The input is a json file with an array of mappings:
//
//  [{"alias":"ml",
//    "value":"mlx.hpc.uio.no"},
//   ...]
//
// The alias mapper can be reinitialized after creation (reading from the same file, which is
// presumed to have changed).  Reinitialization is thread-safe, and if it fails to read the file the
// mapper is unchanged.

package alias

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// MT: Locked
type Aliases struct {
	lock     sync.RWMutex
	filepath string
	mapping  map[string]string
}

func ReadAliases(filepath string) (*Aliases, error) {
	mapping, err := readAliases(filepath)
	if err != nil {
		return nil, err
	}
	return &Aliases{
		filepath: filepath,
		mapping:  mapping,
	}, nil
}

func readAliases(filepath string) (map[string]string, error) {
	all, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
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

	return mapping, nil
}

func (a *Aliases) Resolve(alias string) string {
	a.lock.RLock()
	defer a.lock.RUnlock()
	if m, found := a.mapping[alias]; found {
		return m
	}
	return alias
}

// Map from canonical cluster name to all its aliases.  The map and alias slices are both new.

func (a *Aliases) ReverseExpand() map[string][]string {
	xs := make(map[string][]string)
	a.lock.Lock()
	defer a.lock.Unlock()
	for alias, cluster := range a.mapping {
		xs[cluster] = append(xs[cluster], alias)
	}
	return xs
}

func (a *Aliases) Reread() error {
	mapping, err := readAliases(a.filepath)
	if err != nil {
		return err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	a.mapping = mapping
	return nil
}
