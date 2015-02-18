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

// Check a bool in the database
func GetBool(column, table, row string, id uint) (boolean bool) {

	// Check if thread is closed and get the total amount of posts
	err := db.QueryRow("SELECT ? FROM ? WHERE ? = ?", column, table, row, id).Scan(&boolean)
	if err != nil {
		return false
	}

	return

}

// Set a bool in the database
func SetBool(table, column, row string, boolean bool, id uint) (err error) {

	ps, err = db.Prepare("UPDATE ? SET ?=? WHERE ?=?")
	if err != nil {
		return
	}
	defer ps.Close()

	_, err = updatestatus.Exec(table, column, boolean, row, id)
	if err != nil {
		return
	}

	return

}
