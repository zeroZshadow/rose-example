package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

//GlobalConfig loaded configuration
var GlobalConfig *Config

// Config describes the whole process of generating sitemap
type Config struct {
	Address       string `json:"address"`
	MasterAddress string `json:"masteraddress"`
	Region        string `json:"region"`
	Database      string `json:"database"`
	Name          string `json:"name"`
}

// New create new Config with default values
func New() *Config {
	return &Config{
		Address:       ":0",
		MasterAddress: "ws://localhost:8080/cluster",
		Region:        "EU",
		Database:      "user:password@tcp(localhost.net:3306)/game_db?charset=utf8&parseTime=true",
		Name:          "GameServer1",
	}
}

// FromFile parses file into a Config
func (cfg *Config) FromFile(file string) error {
	path, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	fileC, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(fileC, cfg)
	return err
}
