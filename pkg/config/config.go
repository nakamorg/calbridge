package config

import (
	"encoding/json"
	"os"
)

type config struct {
	Users []User `json:"users"`
}

// loadConfig reads and returns the json config
func loadConfig(path string) (config, error) {
	var conf config
	data, err := os.ReadFile(path)
	if err != nil {
		return conf, err
	}
	err = json.Unmarshal(data, &conf)
	return conf, err
}
