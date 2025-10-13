package cmd

import (
	"errors"

	"sonalyze/data/cluster"
	"sonalyze/db"
	"sonalyze/db/special"
)

func OpenDataStoreFromCommand(anyCmd Command) (err error) {
	db.SetCacheSize(anyCmd.CacheSize())
	if jd := anyCmd.JobanalyzerDir(); jd != "" {
		err = special.OpenFullDataStore(jd)
	} else if dd := anyCmd.DataDir(); dd != "" {
		err = special.OpenDataStoreFromDataDir(dd, anyCmd.ConfigFile())
	} else if rd := anyCmd.ReportDir(); rd != "" {
		err = special.OpenDataStoreFromReportDir(rd, anyCmd.ConfigFile())
	} else if fl := anyCmd.LogFiles(); len(fl) > 0 {
		err = special.OpenDataStoreFromLogFiles(fl, anyCmd.ConfigFile())
	} else if cf := anyCmd.ConfigFile(); cf != "" {
		err = special.OpenDataStoreFromConfigFile(cf)
	} else if anyCmd.Dataless() {
	} else {
		err = errors.New("No data source")
	}
	return
}

func NewMetaFromCommand(anyCmd Command) special.ClusterMeta {
	c := special.LookupCluster(anyCmd.ClusterName())
	if c == nil {
		panic("Cluster name must be defined at this point, could not find '" + anyCmd.ClusterName() + "'")
	}
	return cluster.NewMetaFromCluster(c)
}
