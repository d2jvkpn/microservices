package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const (
	_BcryptCost = 12
)

var (
	_DB *gorm.DB
)

func Connect(dsn string, debug bool) (db *gorm.DB, err error) {
	conf := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	}

	if db, err = gorm.Open(postgres.Open(dsn), conf); err != nil {
		return nil, err
	}

	if debug {
		db = db.Debug()
	}
	_DB = db

	return db, nil
}
