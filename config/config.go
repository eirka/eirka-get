package config

import (
	"encoding/json"
	"os"
)

var Settings *Config

type Config struct {
	Database struct {
		// Database connection settings
		DbUser           string
		DbPassword       string
		DbProto          string
		DbHost           string
		DbDatabase       string
		DbMaxIdle        int
		DbMaxConnections int
	}

	Redis struct {
		// Redis address and max pool connections
		RedisProtocol       string
		RedisAddress        string
		RedisMaxIdle        int
		RedisMaxConnections int
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
