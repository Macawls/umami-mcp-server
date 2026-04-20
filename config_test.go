package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid_env_vars",
			envVars: map[string]string{
				"UMAMI_URL":      "https://test.com",
				"UMAMI_USERNAME": "user",
				"UMAMI_PASSWORD": "pass",
			},
			wantErr: false,
		},
		{
			name: "missing_url",
			envVars: map[string]string{
				"UMAMI_USERNAME": "user",
				"UMAMI_PASSWORD": "pass",
			},
			wantErr: true,
		},
		{
			name:    "missing_all",
			envVars: map[string]string{},
			wantErr: true,
		},
		{
			name: "valid_api_key",
			envVars: map[string]string{
				"UMAMI_URL":     "https://api.umami.is",
				"UMAMI_API_KEY": "secret-key",
			},
			wantErr: false,
		},
		{
			name: "missing_auth",
			envVars: map[string]string{
				"UMAMI_URL": "https://test.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			_, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig_TeamID(t *testing.T) {
	t.Setenv("UMAMI_URL", "https://test.com")
	t.Setenv("UMAMI_USERNAME", "user")
	t.Setenv("UMAMI_PASSWORD", "pass")
	t.Setenv("UMAMI_TEAM_ID", "my-team-123")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	if config.TeamID != "my-team-123" {
		t.Errorf("Expected TeamID 'my-team-123', got '%s'", config.TeamID)
	}
}

func TestLoadConfig_TeamIDOptional(t *testing.T) {
	t.Setenv("UMAMI_URL", "https://test.com")
	t.Setenv("UMAMI_USERNAME", "user")
	t.Setenv("UMAMI_PASSWORD", "pass")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() unexpected error: %v", err)
	}

	if config.TeamID != "" {
		t.Errorf("Expected empty TeamID, got '%s'", config.TeamID)
	}
}
