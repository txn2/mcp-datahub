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
		datahubClient.Close()
	}
}

func TestNewWithInvalidEnvConfig(t *testing.T) {
	// Save and clear env vars
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")
	origTimeout := os.Getenv("DATAHUB_TIMEOUT")

	os.Unsetenv("DATAHUB_URL")
	os.Unsetenv("DATAHUB_TOKEN")
	os.Setenv("DATAHUB_TIMEOUT", "invalid")

	t.Cleanup(func() {
		setEnvOrUnset("DATAHUB_URL", origURL)
		setEnvOrUnset("DATAHUB_TOKEN", origToken)
		setEnvOrUnset("DATAHUB_TIMEOUT", origTimeout)
	})

	opts := DefaultOptions()
	_, _, err := New(opts)
	if err == nil {
		t.Error("New() expected error for invalid env config")
	}
}

func TestNewWithMissingConfig(t *testing.T) {
	// Save and clear env vars
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")

	os.Unsetenv("DATAHUB_URL")
	os.Unsetenv("DATAHUB_TOKEN")

	t.Cleanup(func() {
		setEnvOrUnset("DATAHUB_URL", origURL)
		setEnvOrUnset("DATAHUB_TOKEN", origToken)
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
		datahubClient.Close()
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func setEnvOrUnset(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}
