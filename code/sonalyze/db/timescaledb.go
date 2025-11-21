package db

import (
	"errors"
	"sonalyze/db/types"
)

type ConnectedDB struct{}

func OpenDatabaseURI(databaseURI string) (*ConnectedDB, error) {
	return nil, errors.New("No database connection yet")
}

func (cdb *ConnectedDB) EnumerateClusters() ([]string, error) {
	return nil, errors.New("Database connection not open")
}

func OpenConnectedDB(meta types.Context) (AppendablePersistentDataProvider, error) {
	db := meta.ConnectedDB().(*ConnectedDB)
	_ = db
	return nil, errors.New("Don't know how to open a database yet")
}
