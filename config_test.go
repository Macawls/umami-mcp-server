package main

import (
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set config path to a non-existent file to ensure tests only use env vars
	// This prevents the test from accidentally reading a config.yaml file
	// that might exist in the executable directory
	t.Setenv("UMAMI_MCP_CONFIG", filepath.Join(t.TempDir(), "non-existent-config.yaml"))

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
			name:    "missing_url",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first to ensure clean test state
			t.Setenv("UMAMI_URL", "")
			t.Setenv("UMAMI_USERNAME", "")
			t.Setenv("UMAMI_PASSWORD", "")
			
			// Set test-specific env vars
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
