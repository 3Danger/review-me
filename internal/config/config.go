package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
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
	Jira   Jira   `yaml:"jira"`
	Gitlab Gitlab `yaml:"gitlab"`
}

type Jira struct {
	BaseURL string `yaml:"baseURL"`
	User    string `yaml:"user" required:"true"`
	Pass    string `yaml:"password" required:"true"`
}

type Gitlab struct {
	BaseURL string `yaml:"baseURL"`
	Token   string `yaml:"token" required:"true"`
}

func Load(fileName string) (*Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting current working directory: %w", err)
	}

	file, err := os.Open(filepath.Join(cwd, fileName))
	if err != nil {
		return nil, fmt.Errorf("openning config file: %w", err)
	}
	defer file.Close()

	var c Config
	if err := yaml.NewDecoder(file).Decode(&c); err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	return &c, nil
}
