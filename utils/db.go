package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// NewDb initializes a connection to MySQL and tries to connect.
func NewDb(dbuser, dbpassword, dbproto, dbhost, dbdatabase string, dbmaxidle, dbmaxconnections int) {
	var err error

	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s)/%s", dbuser, dbpassword, dbproto, dbhost, dbdatabase))
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(dbmaxidle)
	db.SetMaxOpenConns(dbmaxconnections)
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
