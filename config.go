package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	UmamiURL string `yaml:"umami_url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}

	configPath := os.Getenv("UMAMI_MCP_CONFIG")
	if configPath == "" {
		exePath, _ := os.Executable()
		configPath = filepath.Join(filepath.Dir(exePath), "config.yaml")
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("invalid config file: %w", err)
		}
	}

	if url := os.Getenv("UMAMI_URL"); url != "" {
		config.UmamiURL = url
	}
	if username := os.Getenv("UMAMI_USERNAME"); username != "" {
		config.Username = username
	}
	if password := os.Getenv("UMAMI_PASSWORD"); password != "" {
		config.Password = password
	}

	if config.UmamiURL == "" || config.Username == "" || config.Password == "" {
		return nil, fmt.Errorf("missing required configuration: UMAMI_URL, UMAMI_USERNAME, UMAMI_PASSWORD")
	}

	return config, nil
}