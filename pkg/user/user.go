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

type Config struct {
	Users []User `json:"users"`
}

// LoadFromJson reads a json file containing calbridge user information and returns the Users
func LoadFromJson(path string) ([]User, error) {
	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}
	return config.Users, err
}

// LoadFromJson reads a json file containing calbridge user information and returns the Users
func LoadConfig(path string) (Config, error) {
	var config Config
	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(data, &config)
	return config, err
}

// LoadFromDB reads calbridge users info from db and returns
func LoadFromDB(path string) ([]User, error) {
	var users []User
	return users, nil
}
