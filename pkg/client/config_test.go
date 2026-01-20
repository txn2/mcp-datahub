package client

import (
	"os"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Timeout != 30*time.Second {
		t.Errorf("DefaultConfig() Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
	if cfg.RetryMax != 3 {
		t.Errorf("DefaultConfig() RetryMax = %v, want %v", cfg.RetryMax, 3)
	}
	if cfg.DefaultLimit != 10 {
		t.Errorf("DefaultConfig() DefaultLimit = %v, want %v", cfg.DefaultLimit, 10)
	}
	if cfg.MaxLimit != 100 {
		t.Errorf("DefaultConfig() MaxLimit = %v, want %v", cfg.MaxLimit, 100)
	}
	if cfg.MaxLineageDepth != 5 {
		t.Errorf("DefaultConfig() MaxLineageDepth = %v, want %v", cfg.MaxLineageDepth, 5)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				URL:   "https://datahub.example.com",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: Config{
				Token: "test-token",
			},
			wantErr: true,
			errMsg:  "DATAHUB_URL is required",
		},
		{
			name: "missing token",
			config: Config{
				URL: "https://datahub.example.com",
			},
			wantErr: true,
			errMsg:  "DATAHUB_TOKEN is required",
		},
		{
			name:    "both missing",
			config:  Config{},
			wantErr: true,
			errMsg:  "DATAHUB_URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

// envVars is a map of environment variable names to values for test setup.
type envVars map[string]string

// setupEnv sets environment variables from the map and clears any not specified.
func setupEnv(t *testing.T, vars envVars) {
	t.Helper()
	allVars := []string{
		"DATAHUB_URL", "DATAHUB_TOKEN", "DATAHUB_TIMEOUT",
		"DATAHUB_RETRY_MAX", "DATAHUB_DEFAULT_LIMIT",
		"DATAHUB_MAX_LIMIT", "DATAHUB_MAX_LINEAGE_DEPTH",
	}
	for _, key := range allVars {
		if val, ok := vars[key]; ok {
			mustSetenv(t, key, val)
		} else {
			mustUnsetenv(t, key)
		}
	}
}

func TestFromEnv(t *testing.T) {
	// Save and restore original env vars
	origEnv := saveEnvVars()
	t.Cleanup(func() { restoreEnvVars(t, origEnv) })

	t.Run("reads basic env vars", func(t *testing.T) {
		setupEnv(t, envVars{
			"DATAHUB_URL":   "https://test.datahub.io",
			"DATAHUB_TOKEN": "test-token-123",
		})

		cfg, err := FromEnv()
		if err != nil {
			t.Fatalf("FromEnv() unexpected error: %v", err)
		}

		if cfg.URL != "https://test.datahub.io" {
			t.Errorf("URL = %v, want %v", cfg.URL, "https://test.datahub.io")
		}
		if cfg.Token != "test-token-123" {
			t.Errorf("Token = %v, want %v", cfg.Token, "test-token-123")
		}
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
		}
	})

	t.Run("reads all env vars", func(t *testing.T) {
		setupEnv(t, envVars{
			"DATAHUB_URL":               "https://full.datahub.io",
			"DATAHUB_TOKEN":             "full-token",
			"DATAHUB_TIMEOUT":           "60",
			"DATAHUB_RETRY_MAX":         "5",
			"DATAHUB_DEFAULT_LIMIT":     "20",
			"DATAHUB_MAX_LIMIT":         "200",
			"DATAHUB_MAX_LINEAGE_DEPTH": "10",
		})

		cfg, err := FromEnv()
		if err != nil {
			t.Fatalf("FromEnv() unexpected error: %v", err)
		}

		if cfg.Timeout != 60*time.Second {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 60*time.Second)
		}
		if cfg.RetryMax != 5 {
			t.Errorf("RetryMax = %v, want %v", cfg.RetryMax, 5)
		}
		if cfg.DefaultLimit != 20 {
			t.Errorf("DefaultLimit = %v, want %v", cfg.DefaultLimit, 20)
		}
		if cfg.MaxLimit != 200 {
			t.Errorf("MaxLimit = %v, want %v", cfg.MaxLimit, 200)
		}
		if cfg.MaxLineageDepth != 10 {
			t.Errorf("MaxLineageDepth = %v, want %v", cfg.MaxLineageDepth, 10)
		}
	})
}

func TestFromEnvInvalidValues(t *testing.T) {
	// Save and restore original env vars
	origEnv := saveEnvVars()
	t.Cleanup(func() { restoreEnvVars(t, origEnv) })

	tests := []struct {
		name string
		vars envVars
	}{
		{
			name: "invalid timeout",
			vars: envVars{
				"DATAHUB_URL":     "https://test.io",
				"DATAHUB_TOKEN":   "token",
				"DATAHUB_TIMEOUT": "invalid",
			},
		},
		{
			name: "invalid retry max",
			vars: envVars{
				"DATAHUB_URL":       "https://test.io",
				"DATAHUB_TOKEN":     "token",
				"DATAHUB_RETRY_MAX": "not-a-number",
			},
		},
		{
			name: "invalid default limit",
			vars: envVars{
				"DATAHUB_URL":           "https://test.io",
				"DATAHUB_TOKEN":         "token",
				"DATAHUB_DEFAULT_LIMIT": "abc",
			},
		},
		{
			name: "invalid max limit",
			vars: envVars{
				"DATAHUB_URL":       "https://test.io",
				"DATAHUB_TOKEN":     "token",
				"DATAHUB_MAX_LIMIT": "xyz",
			},
		},
		{
			name: "invalid max lineage depth",
			vars: envVars{
				"DATAHUB_URL":               "https://test.io",
				"DATAHUB_TOKEN":             "token",
				"DATAHUB_MAX_LINEAGE_DEPTH": "bad",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupEnv(t, tt.vars)
			_, err := FromEnv()
			if err == nil {
				t.Errorf("FromEnv() expected error for %s", tt.name)
			}
		})
	}
}

func mustSetenv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set %s: %v", key, err)
	}
}

func mustUnsetenv(t *testing.T, key string) {
	t.Helper()
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset %s: %v", key, err)
	}
}

// saveEnvVars saves the current values of all DATAHUB_* environment variables.
func saveEnvVars() map[string]string {
	keys := []string{
		"DATAHUB_URL", "DATAHUB_TOKEN", "DATAHUB_TIMEOUT",
		"DATAHUB_RETRY_MAX", "DATAHUB_DEFAULT_LIMIT",
		"DATAHUB_MAX_LIMIT", "DATAHUB_MAX_LINEAGE_DEPTH",
	}
	saved := make(map[string]string)
	for _, k := range keys {
		saved[k] = os.Getenv(k)
	}
	return saved
}

// restoreEnvVars restores environment variables from a saved map.
func restoreEnvVars(t *testing.T, saved map[string]string) {
	t.Helper()
	for k, v := range saved {
		if v == "" {
			mustUnsetenv(t, k)
		} else {
			mustSetenv(t, k, v)
		}
	}
}
