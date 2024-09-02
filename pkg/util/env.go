package util

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv loads environment variables from a specified .env file.
// If filePath is empty, it tries to load the .env from pwd
func LoadDotEnv(filePath string) error {
	if filePath == "" {
		filePath = ".env"
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
	return scanner.Err()
}
