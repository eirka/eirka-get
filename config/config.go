package config

import (
	"encoding/json"
	"fmt"
	"os"
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
