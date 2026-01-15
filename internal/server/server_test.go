package server

import (
	"os"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.MultiServerConfig != nil {
		t.Error("DefaultOptions() MultiServerConfig should be nil")
	}
	if opts.ToolkitConfig.DefaultLimit != 10 {
		t.Errorf("DefaultOptions() ToolkitConfig.DefaultLimit = %d, want 10", opts.ToolkitConfig.DefaultLimit)
	}
}

func TestNewWithProvidedConfig(t *testing.T) {
	cfg := &multiserver.Config{
		Default: "datahub",
		Primary: client.Config{
			URL:   "https://test.datahub.io",
			Token: "test-token",
		},
		Connections: map[string]multiserver.ConnectionConfig{},
	}

	opts := Options{
		MultiServerConfig: cfg,
	}

	server, mgr, err := New(opts)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if server == nil {
		t.Error("New() returned nil server")
	}
	if mgr == nil {
		t.Error("New() returned nil manager")
	}

	// Check that we have the expected connection
	if !mgr.HasConnection("datahub") {
		t.Error("Manager should have 'datahub' connection")
	}

	// Clean up
	if mgr != nil {
		if err := mgr.Close(); err != nil {
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
	server, mgr, err := New(opts)

	// Should not error - server starts even without config
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if server == nil {
		t.Error("New() returned nil server even in unconfigured mode")
	}
	if mgr == nil {
		t.Error("New() returned nil manager even in unconfigured mode")
	}

	// Manager should still be valid, just with no valid client configs
	if mgr != nil {
		if err := mgr.Close(); err != nil {
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
