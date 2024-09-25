// The "config" table
//
// The config objects are cached; only explicit invalidation will flush the cache.

package db

import (
	"sync"

	"go-utils/config"
)

var (
	// MT: Locked + contained objects are immutable after creation
	configCacheLock sync.Mutex
	configCache     = make(map[string]*config.ClusterConfig)
)

// Read config if the name is not empty

func MaybeGetConfig(configName string) (*config.ClusterConfig, error) {
	if configName == "" {
		return nil, nil
	}
	return ReadConfigData(configName)
}

// This reads or returns the cached config, which is a shared object that will not be modified
// subsequently and must not be modified by the caller.  (Cache invalidation will never invalidate
// the object either, but subsequent calls may return a different object.)

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

func InvalidateConfigCache() {
	configCacheLock.Lock()
	defer configCacheLock.Unlock()

	clear(configCache)
}
