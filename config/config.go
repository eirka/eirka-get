package config

import (
	"encoding/json"
	"fmt"
	"os"
)

func init() {
	file, err := os.Open("/etc/pram/pram.conf")
	if err != nil {
		// file was not found so use default settings
		Settings = &Config{
			Get: Get{
				Host: "127.0.0.1",
				Port: 5010,
			},
		}
		return
	}

	// if the file is found fill settings with json
	Settings = &Config{}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// Settings holds the current config options
var Settings *Config

// Config represents the possible configurable parameters
// for the local daemon
type Config struct {
	Get      Get
	CORS     CORS
	Database Database
	Redis    Redis
}

// Get sets what the daemon listens on
type Get struct {
	Host                   string
	Port                   uint
	DatabaseMaxIdle        int
	DatabaseMaxConnections int
	RedisMaxIdle           int
	RedisMaxConnections    int
	DataDog                bool
}

// Database holds the connection settings for MySQL
type Database struct {
	Host     string
	Protocol string
	User     string
	Password string
	Database string
}

// Redis holds the connection settings for the redis cache
type Redis struct {
	Host     string
	Protocol string
}

// CORS is a list of allowed remote addresses
type CORS struct {
	Sites []string
}
