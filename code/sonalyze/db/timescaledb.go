package db

import (
	"errors"
	"sonalyze/db/special"
)

func OpenConnectedDB(meta special.ClusterMeta) (DataProvider, error) {
	return nil, errors.New("Don't know how to open a database yet")
}
