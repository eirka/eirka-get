package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Settings holds the current config options
var Settings *Config

// Config represents the possible configurable parameters
type Config struct {
	Get struct {
		// Settings for daemon
		Address string
		Port    uint
	}

	// sites for CORS
	CORS struct {
		Sites []string
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

	// HMAC secret for bcrypt
	Session struct {
		Secret string
	}
}

func init() {
	file, err := os.Open("/etc/pram/pram.conf")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	Settings = &Config{}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&Settings)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
