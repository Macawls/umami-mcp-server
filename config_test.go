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
