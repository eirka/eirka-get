package config

import (
	"encoding/json"
	"fmt"
	"os"

	u "github.com/techjanitor/pram-post/utils"
)

var (
	PramVersion = "1.1.2"
	Settings    *Config
)

type Config struct {
	General struct {
		// Settings for daemon
		Address string
		Port    uint
	}

	Database struct {
		// Database connection settings
		User           string
		Password       string
		Proto          string
		Host           string
		Database       string
		MaxIdle        int
		MaxConnections int
	}

	Redis struct {
		// Redis address and max pool connections
		Protocol       string
		Address        string
		MaxIdle        int
		MaxConnections int
	}

	Antispam struct {
		// Antispam cookie
		CookieName  string
		CookieValue string
	}

	Limits struct {
		// Set default posts per page
		PostsPerPage uint

		// Set default threads per index page
		ThreadsPerPage uint
		// Add one to number because first post is included
		PostsPerThread uint

		// Max request parameter input size
		ParamMaxSize uint
	}
}

// Prints the current config during start if debug
func Print() {

	// Marshal the structs into JSON
	output, err := json.MarshalIndent(Settings, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", output)

	return

}

// Get the config file and decode into struct
func init() {
	file, err := os.Open("/etc/pram/get.conf")
	if err != nil {
		panic(err)
	}

	Settings = &Config{}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Settings)
	if err != nil {
		panic(err)
	}

}

// Get limits that are in the database
func GetDatabaseSettings() {

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		panic(err)
	}

	ps, err := db.Prepare("SELECT settings_value FROM settings WHERE settings_key = ? LIMIT 1")
	if err != nil {
		panic(err)
	}
	defer ps.Close()

	err = ps.QueryRow("antispam_cookiename").Scan(&Settings.Antispam.CookieName)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("antispam_cookievalue").Scan(&Settings.Antispam.CookieValue)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("thread_postsperpage").Scan(&Settings.Limits.PostsPerPage)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("index_threadsperpage").Scan(&Settings.Limits.ThreadsPerPage)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("index_postsperthread").Scan(&Settings.Limits.PostsPerThread)
	if err != nil {
		panic(err)
	}

	err = ps.QueryRow("param_maxsize").Scan(&Settings.Limits.ParamMaxSize)
	if err != nil {
		panic(err)
	}

	return

}
