package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// locationCache caches loaded time locations to avoid repeated disk I/O.
var locationCache sync.Map

// LoadLocation returns a cached *time.Location for the given timezone name.
// It only loads from the IANA database on the first request for each timezone.
func LoadLocation(tz string) *time.Location {
	if cached, ok := locationCache.Load(tz); ok {
		return cached.(*time.Location)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	locationCache.Store(tz, loc)
	return loc
}

type Config struct {
	Jira    Jira
	Gitlab  Gitlab
	Message Message
}

type Message struct {
	Team   string
	Review string
	Deploy string
}

type Jira struct {
	BaseURL       string
	User          string
	Pass          string
	ProjectPrefix string
}

type Gitlab struct {
	BaseURL string
	Token   string
}

// envOrDefault returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or empty.
func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func Load() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current working directory: %w", err)
	}

	// Load .env file if it exists; silently skip if not
	envPath := filepath.Join(cwd, ".env")
	if err := godotenv.Load(envPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("loading .env file: %w", err)
	}

	c := &Config{
		Jira: Jira{
			BaseURL:       os.Getenv("JIRA_BASE_URL"),
			User:          os.Getenv("JIRA_USER"),
			Pass:          os.Getenv("JIRA_PASSWORD"),
			ProjectPrefix: envOrDefault("JIRA_PROJECT_PREFIX", "FD-"),
		},
		Gitlab: Gitlab{
			BaseURL: os.Getenv("GITLAB_BASE_URL"),
			Token:   os.Getenv("GITLAB_TOKEN"),
		},
		Message: Message{
			Team:   os.Getenv("MESSAGE_TEAM"),
			Review: os.Getenv("MESSAGE_REVIEW"),
			Deploy: os.Getenv("MESSAGE_DEPLOY"),
		},
	}

	return c, nil
}
