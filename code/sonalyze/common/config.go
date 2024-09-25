package common

import (
	"sync"

	"go-utils/config"
)

const (
	ClusterAliasesFilename = "cluster-aliases.json"
)

// The config objects are cached.  Currently this cache is never cleaned.  It could be cleaned if
// the daemon code catches SIGHUP (telling it to reinitialize itself), but in that case we must also
// purge the entire LogFile cache, as those data have been rectified with config data.

var (
	// MT: Locked
	configCacheLock sync.Mutex
	configCache     = make(map[string]*config.ClusterConfig)
)

func GetConfig(configName string) (*config.ClusterConfig, error) {
	configCacheLock.Lock()
	defer configCacheLock.Unlock()

	if probe := configCache[configName]; probe != nil {
		return probe, nil
	}

	cfg, err := config.ReadConfig(configName)
	if err != nil {
		return nil, err
	}
	configCache[configName] = cfg
	return cfg, nil
}

func MaybeGetConfig(configName string) (*config.ClusterConfig, error) {
	if configName == "" {
		return nil, nil
	}
	return GetConfig(configName)
}
