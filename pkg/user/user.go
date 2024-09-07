package user

import (
	"encoding/json"
	"os"
)

type User struct {
	Name   string `json:"name"`
	CalDAV struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"caldav"`
	SMTP struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"smtp"`
	IMAP struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"imap"`
}

type config struct {
	Users []User `json:"users"`
}

// LoadFromJson reads a json file containing calbridge user information and returns the Users
func LoadFromJson(path string) ([]User, error) {
	config, err := loadConfig(path)
	if err != nil {
		return nil, err
	}
	return config.Users, err
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

// LoadFromDB reads calbridge users info from db and returns
func LoadFromDB(path string) ([]User, error) {
	var users []User
	return users, nil
}
