package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"

	"github.com/techjanitor/pram-get/config"
)

var db *sql.DB

// NewDb initializes a connection to MySQL and tries to connect.
func NewDb() {
	var err error

	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s)/%s",
		config.Settings.Database.User,
		config.Settings.Database.Password,
		config.Settings.Database.Proto,
		config.Settings.Database.Host,
		config.Settings.Database.Database,
	))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(config.Settings.Database.MaxIdle)
	db.SetMaxOpenConns(config.Settings.Database.MaxConnections)
}

// CloseDb closes the connection to MySQL
func CloseDb() (err error) {
	return db.Close()
}

// GetDb returns a connection to MySQL
func GetDb() (*sql.DB, error) {
	return db, nil
}

// GetTransaction will return a transaction
func GetTransaction() (*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	return tx, err
}
