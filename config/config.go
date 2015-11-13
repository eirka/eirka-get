package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

var Settings *Config

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
}

func Print() {

	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n", "Local Config")
	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n\n", "Server")
	fmt.Printf("%-20v%40v\n", "Type", "GET")
	fmt.Printf("%-20v%40v\n", "Address", Settings.Get.Address)
	fmt.Printf("%-20v%40v\n", "Port", Settings.Get.Port)
	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n\n", "CORS")
	fmt.Printf("%-20v%40v\n", "Domains", strings.Join(Settings.CORS.Sites, ", "))
	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n\n", "Database")
	fmt.Printf("%-20v%40v\n", "User", Settings.Database.User)
	fmt.Printf("%-20v%40v\n", "Password", Settings.Database.Password)
	fmt.Printf("%-20v%40v\n", "Protocol", Settings.Database.Proto)
	fmt.Printf("%-20v%40v\n", "Host", Settings.Database.Host)
	fmt.Printf("%-20v%40v\n", "Database", Settings.Database.Database)
	fmt.Printf("%-20v%40v\n", "Max Idle", Settings.Database.MaxIdle)
	fmt.Printf("%-20v%40v\n", "Max Connections", Settings.Database.MaxConnections)
	fmt.Println(strings.Repeat("*", 60))
	fmt.Printf("%-20v\n\n", "Redis")
	fmt.Printf("%-20v%40v\n", "Protocol", Settings.Redis.Protocol)
	fmt.Printf("%-20v%40v\n", "Address", Settings.Redis.Address)
	fmt.Printf("%-20v%40v\n", "Max Idle", Settings.Redis.MaxIdle)
	fmt.Printf("%-20v%40v\n", "Max Connections", Settings.Redis.MaxConnections)
	fmt.Println(strings.Repeat("*", 60))

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
