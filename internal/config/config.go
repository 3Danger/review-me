package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var timeLocationMSK *time.Location

func init() {
	var err error
	timeLocationMSK, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic(fmt.Errorf("loading Moscow time location: %w", err))
	}
}

func LocationMSK() *time.Location {
	return timeLocationMSK
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
	BaseURL string
	User    string
	Pass    string
}

type Gitlab struct {
	BaseURL string
	Token   string
}

func Load() (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current working directory: %w", err)
	}

	envMap, err := readEnvFile(filepath.Join(cwd, ".env"))
	if err != nil {
		return nil, fmt.Errorf("reading .env file: %w", err)
	}

	c := &Config{
		Jira: Jira{
			BaseURL: lookup("JIRA_BASE_URL", envMap),
			User:    lookup("JIRA_USER", envMap),
			Pass:    lookup("JIRA_PASSWORD", envMap),
		},
		Gitlab: Gitlab{
			BaseURL: lookup("GITLAB_BASE_URL", envMap),
			Token:   lookup("GITLAB_TOKEN", envMap),
		},
		Message: Message{
			Team:   lookup("MESSAGE_TEAM", envMap),
			Review: lookup("MESSAGE_REVIEW", envMap),
			Deploy: lookup("MESSAGE_DEPLOY", envMap),
		},
	}

	return c, nil
}

func lookup(key string, envMap map[string]string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return envMap[key]
}

func readEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		result[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
