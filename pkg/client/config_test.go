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

func TestFromEnv(t *testing.T) {
	// Save original env vars
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")
	origTimeout := os.Getenv("DATAHUB_TIMEOUT")
	origRetryMax := os.Getenv("DATAHUB_RETRY_MAX")
	origDefaultLimit := os.Getenv("DATAHUB_DEFAULT_LIMIT")
	origMaxLimit := os.Getenv("DATAHUB_MAX_LIMIT")
	origMaxDepth := os.Getenv("DATAHUB_MAX_LINEAGE_DEPTH")

	// Restore on cleanup
	t.Cleanup(func() {
		setEnvOrUnsetT(t, "DATAHUB_URL", origURL)
		setEnvOrUnsetT(t, "DATAHUB_TOKEN", origToken)
		setEnvOrUnsetT(t, "DATAHUB_TIMEOUT", origTimeout)
		setEnvOrUnsetT(t, "DATAHUB_RETRY_MAX", origRetryMax)
		setEnvOrUnsetT(t, "DATAHUB_DEFAULT_LIMIT", origDefaultLimit)
		setEnvOrUnsetT(t, "DATAHUB_MAX_LIMIT", origMaxLimit)
		setEnvOrUnsetT(t, "DATAHUB_MAX_LINEAGE_DEPTH", origMaxDepth)
	})

	t.Run("reads basic env vars", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://test.datahub.io")
		mustSetenv(t, "DATAHUB_TOKEN", "test-token-123")
		mustUnsetenv(t, "DATAHUB_TIMEOUT")
		mustUnsetenv(t, "DATAHUB_RETRY_MAX")
		mustUnsetenv(t, "DATAHUB_DEFAULT_LIMIT")
		mustUnsetenv(t, "DATAHUB_MAX_LIMIT")
		mustUnsetenv(t, "DATAHUB_MAX_LINEAGE_DEPTH")

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
		// Should have defaults
		if cfg.Timeout != 30*time.Second {
			t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
		}
	})

	t.Run("reads all env vars", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://full.datahub.io")
		mustSetenv(t, "DATAHUB_TOKEN", "full-token")
		mustSetenv(t, "DATAHUB_TIMEOUT", "60")
		mustSetenv(t, "DATAHUB_RETRY_MAX", "5")
		mustSetenv(t, "DATAHUB_DEFAULT_LIMIT", "20")
		mustSetenv(t, "DATAHUB_MAX_LIMIT", "200")
		mustSetenv(t, "DATAHUB_MAX_LINEAGE_DEPTH", "10")

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

	t.Run("invalid timeout", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://test.io")
		mustSetenv(t, "DATAHUB_TOKEN", "token")
		mustSetenv(t, "DATAHUB_TIMEOUT", "invalid")

		_, err := FromEnv()
		if err == nil {
			t.Error("FromEnv() expected error for invalid timeout")
		}
	})

	t.Run("invalid retry max", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://test.io")
		mustSetenv(t, "DATAHUB_TOKEN", "token")
		mustUnsetenv(t, "DATAHUB_TIMEOUT")
		mustSetenv(t, "DATAHUB_RETRY_MAX", "not-a-number")

		_, err := FromEnv()
		if err == nil {
			t.Error("FromEnv() expected error for invalid retry max")
		}
	})

	t.Run("invalid default limit", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://test.io")
		mustSetenv(t, "DATAHUB_TOKEN", "token")
		mustUnsetenv(t, "DATAHUB_TIMEOUT")
		mustUnsetenv(t, "DATAHUB_RETRY_MAX")
		mustSetenv(t, "DATAHUB_DEFAULT_LIMIT", "abc")

		_, err := FromEnv()
		if err == nil {
			t.Error("FromEnv() expected error for invalid default limit")
		}
	})

	t.Run("invalid max limit", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://test.io")
		mustSetenv(t, "DATAHUB_TOKEN", "token")
		mustUnsetenv(t, "DATAHUB_TIMEOUT")
		mustUnsetenv(t, "DATAHUB_RETRY_MAX")
		mustUnsetenv(t, "DATAHUB_DEFAULT_LIMIT")
		mustSetenv(t, "DATAHUB_MAX_LIMIT", "xyz")

		_, err := FromEnv()
		if err == nil {
			t.Error("FromEnv() expected error for invalid max limit")
		}
	})

	t.Run("invalid max lineage depth", func(t *testing.T) {
		mustSetenv(t, "DATAHUB_URL", "https://test.io")
		mustSetenv(t, "DATAHUB_TOKEN", "token")
		mustUnsetenv(t, "DATAHUB_TIMEOUT")
		mustUnsetenv(t, "DATAHUB_RETRY_MAX")
		mustUnsetenv(t, "DATAHUB_DEFAULT_LIMIT")
		mustUnsetenv(t, "DATAHUB_MAX_LIMIT")
		mustSetenv(t, "DATAHUB_MAX_LINEAGE_DEPTH", "bad")

		_, err := FromEnv()
		if err == nil {
			t.Error("FromEnv() expected error for invalid max lineage depth")
		}
	})
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

func setEnvOrUnsetT(t *testing.T, key, value string) {
	t.Helper()
	var err error
	if value == "" {
		err = os.Unsetenv(key)
	} else {
		err = os.Setenv(key, value)
	}
	if err != nil {
		t.Errorf("failed to set/unset env %s: %v", key, err)
	}
}
