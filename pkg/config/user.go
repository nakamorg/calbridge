package config

type User struct {
	Name string `json:"name"`
	// How often events and emails are checked, parsed as golang time. Ex: 30m, 1h, 3h etc
	Frequency string `json:"frequency"`
	CalDAV    struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
		// Number of upcoming days for which to read CalDAV events and send invitations.
		EventDays int `json:"eventDays"`
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
		// Number of past hours from which to read emails for calendar invites.
		EmailHours int `json:"emailHours"`
	} `json:"imap"`
}

// LoadFromConfig reads a json file containing calbridge user information and returns the Users
func LoadFromConfig(path string) ([]User, error) {
	config, err := loadConfig(path)
	if err != nil {
		return nil, err
	}
	return config.Users, err
}

// LoadFromDB reads calbridge users info from db and returns
func LoadFromDB(path string) ([]User, error) {
	var users []User
	return users, nil
}
