package cmd

import (
	"go-utils/config"
	"sonalyze/db"
	"sonalyze/db/special"
)

type trivialMeta struct {
	cluster        *special.ClusterEntry
	cfg            *config.ClusterConfig // Temporary
	provider       db.DataProvider
}

func NewMetaFromConfig(cluster *special.ClusterEntry, cfg *config.ClusterConfig) {
	return &trivialMeta{cluster, cfg, nil}
}

// This is temporary!
func (tm *trivialMeta) SetConfig(cfg *config.ClusterConfig) {
	tm.cfg = cfg
}

func (tm *trivialMeta) ClusterName() string {
	return tm.cluster.Name
}

func (tm *trivialMeta) ExcludedUsers() []string {
	return tm.cluster.ExcludeUser
}

func (tm *trivialMeta) SetDataProvider(provider any) {
	if tm.provider != nil {
		panic("Can only set data provider once")
	}
	tm.provider = provider.(db.DataProvider)
}

func (tm *trivialMeta) LookupHostByTime(host string, time int64) *config.NodeConfigRecord {
	if tm.cfg != nil {
		return tm.cfg.LookupHost(host)
	}
	return nil
}

func (tm *trivialMeta) HostsDefinedInTimeWindow(fromIncl, toIncl int64) []string {
	if tm.cfg != nil {
		return tm.cfg.HostsDefinedInTimeWindow(fromIncl, toIncl)
	}
	return nil
}


// This will change

// TODO: This will change radically.  Now that we have a DataProvider we can perform the necessary
// lookup.  But that depends on the provider being the right *kind* of provider.  There needs to be
// a compatibility test here.  This test will usually fail for a FileListDB but should always
// succeed for a PersistentDirectoryDB (and of course for a time series DB).  We should return nil
// for a FileListDB that is not matching, this corresponds to old no-config-file behavior.

// func (tm *trivialMeta) getConfig() *config.ClusterConfig {
// 	cfg, err := special.MaybeGetConfig(tm.configFileName)
// 	if err == nil {
// 		return cfg
// 	}
// 	return nil
// }

const anonCluster = "*"

func NewMetaFromCommand(anyCmd Command) special.ClusterMeta {
	// TODO: And now that we have this in a clean state, we see that ClusterName always will be ""
	// here because it is not provided to us when we run against a -data-dir, in fact -cluster and
	// -data-dir are mutually exclusive and ClusterName() is part of RemotableArgs.  So that's
	// interesting and needs fixing.  For daemon it is different - it explicitly has cluster names.
	//
	// For now, and for here, the cluster name does not matter, so we're OK, but it's just showing
	// how things are creaking.

	var configName string
	if c, ok := anyCmd.(interface{ ConfigFile() string }); ok {
		configName = c.ConfigFile()
	}
	clusterName := anonCluster
	if c, ok := anyCmd.(interface{ ClusterName() string }); ok {
		clusterName = c.ClusterName()
	}
	return NewMetaFromNames(clusterName, configName)
}
