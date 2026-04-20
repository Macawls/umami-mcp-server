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
	APIKey   string `yaml:"api_key"`
	TeamID   string `yaml:"team_id"`
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
	if apiKey := os.Getenv("UMAMI_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}
	if teamID := os.Getenv("UMAMI_TEAM_ID"); teamID != "" {
		config.TeamID = teamID
	}

	if config.UmamiURL == "" {
		return nil, fmt.Errorf("missing required configuration: UMAMI_URL")
	}
	if config.APIKey == "" && (config.Username == "" || config.Password == "") {
		return nil, fmt.Errorf("missing required configuration: set UMAMI_API_KEY " +
			"(for Umami Cloud) or both UMAMI_USERNAME and UMAMI_PASSWORD (for self-hosted)")
	}

	return config, nil
}
