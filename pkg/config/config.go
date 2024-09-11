package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// CreateSampleConfig creates a sample config file at path
func CreateSampleConfig(path string) error {
	config := config{
		Users: []User{
			{
				Name:      "user1",
				Frequency: "1h",
				CalDAV: struct {
					URL       string `json:"url"`
					Username  string `json:"username"`
					Password  string `json:"password"`
					EventDays int    `json:"eventDays"`
				}{
					URL:       "https://caldav.example.com/calendars/user1/xyz/",
					Username:  "user1",
					Password:  "password1",
					EventDays: 5,
				},
				SMTP: struct {
					Host     string `json:"host"`
					Username string `json:"username"`
					Password string `json:"password"`
				}{
					Host:     "mail.example.org",
					Username: "user1@example.org",
					Password: "password1",
				},
				IMAP: struct {
					Host       string `json:"host"`
					Username   string `json:"username"`
					Password   string `json:"password"`
					EmailHours int    `json:"emailHours"`
				}{
					Host:       "mail.example.org",
					Username:   "user1@example.org",
					Password:   "password1",
					EmailHours: 6,
				},
			},
		},
	}
	err := os.MkdirAll(filepath.Dir(path), 0744)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return file.Sync()
}
