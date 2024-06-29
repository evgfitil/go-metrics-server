package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/caarlos0/env/v10"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "default config",
			want: &Config{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name: "valid environment variables",
			envVars: map[string]string{
				"ADDRESS":           "localhost:9090",
				"STORE_INTERVAL":    "600",
				"FILE_STORAGE_PATH": "/tmp/test-db.json",
				"RESTORE":           "true",
				"DATABASE_DSN":      "user:password@/dbname",
			},
			expected: &Config{
				BindAddress:     "localhost:9090",
				StoreInterval:   600,
				FileStoragePath: "/tmp/test-db.json",
				Restore:         true,
				DatabaseDSN:     "user:password@/dbname",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg := NewConfig()
			if err := env.Parse(cfg); err != nil {
				t.Fatalf("Failed to parse env vars: %v", err)
			}

			if !reflect.DeepEqual(cfg, tt.expected) {
				t.Errorf("ConfigFromEnv() = %v, want %v", cfg, tt.expected)
			}

			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
