package db

import (
	"errors"
	"sonalyze/db/types"
)

func OpenConnectedDB(meta types.Context) (DataProvider, error) {
	return nil, errors.New("Don't know how to open a database yet")
}

type ConnectedDB struct{}

func OpenDatabaseURI(databaseURI string) (*ConnectedDB, error) {
	return nil, errors.New("No database connection yet")
}

func (cdb *ConnectedDB) EnumerateClusters() ([]string, error) {
	return nil, errors.New("Database connection not open")
}
