// The "config" table.  This is independent of any underlying database engine, the config file is
// currently always a real file (one per cluster), and does not live in the database per se.
//
// The config objects are cached; only explicit invalidation will flush the cache.

package special

import (
	"sync"

	"go-utils/config"
)

var (
	// MT: Locked + contained objects are immutable after creation
	configCacheLock sync.Mutex
	configCache     = make(map[string]*config.ClusterConfig)
)

// Read the config file if the file name is not empty.
func MaybeGetConfig(configFileName string) (*config.ClusterConfig, error) {
	if configFileName == "" {
		return nil, nil
	}
	return ReadConfigData(configFileName)
}

// ReadConfigData reads or returns the cached config, which is a shared object that will not be
// modified subsequently and must not be modified by the caller.  (Cache invalidation will never
// invalidate the object either, but subsequent calls may return a different object.)
func ReadConfigData(configFileName string) (*config.ClusterConfig, error) {
	configCacheLock.Lock()
	defer configCacheLock.Unlock()

	if probe := configCache[configFileName]; probe != nil {
		return probe, nil
	}

	cfg, err := config.ReadConfig(configFileName)
	if err != nil {
		return nil, err
	}
	configCache[configFileName] = cfg
	return cfg, nil
}

// Invalidate all cached config data from all files.  Data that have been returned earlier continues
// to be live, but ReadConfigData will re-read.
func InvalidateConfigCache() {
	configCacheLock.Lock()
	defer configCacheLock.Unlock()

	clear(configCache)
}
