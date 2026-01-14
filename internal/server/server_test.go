package server

import (
	"os"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/client"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.ClientConfig != nil {
		t.Error("DefaultOptions() ClientConfig should be nil")
	}
	if opts.ToolkitConfig.DefaultLimit != 10 {
		t.Errorf("DefaultOptions() ToolkitConfig.DefaultLimit = %d, want 10", opts.ToolkitConfig.DefaultLimit)
	}
}

func TestNewWithProvidedConfig(t *testing.T) {
	cfg := &client.Config{
		URL:   "https://test.datahub.io",
		Token: "test-token",
	}

	opts := Options{
		ClientConfig: cfg,
	}

	server, datahubClient, err := New(opts)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if server == nil {
		t.Error("New() returned nil server")
	}
	if datahubClient == nil {
		t.Error("New() returned nil client")
	}

	// Clean up
	if datahubClient != nil {
		if err := datahubClient.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	}
}

func TestNewWithInvalidEnvConfig(t *testing.T) {
	// Save original values
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")
	origTimeout := os.Getenv("DATAHUB_TIMEOUT")

	// Set test values
	if err := os.Unsetenv("DATAHUB_URL"); err != nil {
		t.Fatalf("failed to unset DATAHUB_URL: %v", err)
	}
	if err := os.Unsetenv("DATAHUB_TOKEN"); err != nil {
		t.Fatalf("failed to unset DATAHUB_TOKEN: %v", err)
	}
	if err := os.Setenv("DATAHUB_TIMEOUT", "invalid"); err != nil {
		t.Fatalf("failed to set DATAHUB_TIMEOUT: %v", err)
	}

	t.Cleanup(func() {
		restoreEnv(t, "DATAHUB_URL", origURL)
		restoreEnv(t, "DATAHUB_TOKEN", origToken)
		restoreEnv(t, "DATAHUB_TIMEOUT", origTimeout)
	})

	opts := DefaultOptions()
	_, _, err := New(opts)
	if err == nil {
		t.Error("New() expected error for invalid env config")
	}
}

func TestNewWithMissingConfig(t *testing.T) {
	// Save original values
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")

	// Unset values
	if err := os.Unsetenv("DATAHUB_URL"); err != nil {
		t.Fatalf("failed to unset DATAHUB_URL: %v", err)
	}
	if err := os.Unsetenv("DATAHUB_TOKEN"); err != nil {
		t.Fatalf("failed to unset DATAHUB_TOKEN: %v", err)
	}

	t.Cleanup(func() {
		restoreEnv(t, "DATAHUB_URL", origURL)
		restoreEnv(t, "DATAHUB_TOKEN", origToken)
	})

	opts := DefaultOptions()
	server, datahubClient, err := New(opts)

	// Should not error - server starts even without config
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if server == nil {
		t.Error("New() returned nil server even in unconfigured mode")
	}
	if datahubClient != nil {
		t.Error("New() should return nil client when unconfigured")
		if err := datahubClient.Close(); err != nil {
			t.Errorf("Close() error: %v", err)
		}
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func restoreEnv(t *testing.T, key, value string) {
	t.Helper()
	var err error
	if value == "" {
		err = os.Unsetenv(key)
	} else {
		err = os.Setenv(key, value)
	}
	if err != nil {
		t.Errorf("failed to restore env %s: %v", key, err)
	}
}
