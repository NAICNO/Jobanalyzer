package cmd

import (
	"errors"

	"sonalyze/db"
	"sonalyze/db/special"
	"sonalyze/db/types"
)

func OpenDataStoreFromCommand(anyCmd Command) (err error) {
	db.SetCacheSize(anyCmd.CacheSize())
	if jd := anyCmd.JobanalyzerDir(); jd != "" {
		err = db.OpenFullDataStore(jd, anyCmd.DatabaseURI())
	} else if dd := anyCmd.DataDir(); dd != "" {
		err = db.OpenDataStoreFromDataDir(dd, anyCmd.ConfigFile())
	} else if rd := anyCmd.ReportDir(); rd != "" {
		err = db.OpenDataStoreFromReportDir(rd, anyCmd.ConfigFile())
	} else if fl := anyCmd.LogFiles(); len(fl) > 0 {
		err = db.OpenDataStoreFromLogFiles(fl, anyCmd.ConfigFile())
	} else if cf := anyCmd.ConfigFile(); cf != "" {
		err = db.OpenDataStoreFromConfigFile(cf)
	} else if anyCmd.Dataless() {
	} else {
		err = errors.New("No data source")
	}
	return
}

func NewContextFromCommand(anyCmd Command) types.Context {
	c := special.LookupCluster(anyCmd.ClusterName())
	if c == nil {
		panic("Cluster must be defined at this point, could not find '" + anyCmd.ClusterName() + "'")
	}
	return db.NewContextFromCluster(c)
}
