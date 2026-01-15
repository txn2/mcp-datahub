package multiserver

import (
	"testing"
	"time"

	"github.com/txn2/mcp-datahub/pkg/client"
)

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		wantErr     bool
		wantCount   int
		wantDefault string
	}{
		{
			name: "primary only",
			envVars: map[string]string{
				"DATAHUB_URL":   "https://datahub.example.com",
				"DATAHUB_TOKEN": "test-token",
			},
			wantErr:     false,
			wantCount:   1,
			wantDefault: "datahub",
		},
		{
			name: "custom connection name",
			envVars: map[string]string{
				"DATAHUB_URL":             "https://datahub.example.com",
				"DATAHUB_TOKEN":           "test-token",
				"DATAHUB_CONNECTION_NAME": "Production Data",
			},
			wantErr:     false,
			wantCount:   1,
			wantDefault: "Production Data",
		},
		{
			name: "with additional servers",
			envVars: map[string]string{
				"DATAHUB_URL":                "https://prod.datahub.example.com",
				"DATAHUB_TOKEN":              "prod-token",
				"DATAHUB_ADDITIONAL_SERVERS": `{"staging": {"url": "https://staging.datahub.example.com", "token": "staging-token"}}`,
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "multiple additional servers",
			envVars: map[string]string{
				"DATAHUB_URL":   "https://prod.datahub.example.com",
				"DATAHUB_TOKEN": "prod-token",
				"DATAHUB_ADDITIONAL_SERVERS": `{"staging": {"url": "https://staging.datahub.example.com"},` +
					` "dev": {"url": "https://dev.datahub.example.com", "timeout": 60}}`,
			},
			wantErr:   false,
			wantCount: 3,
		},
		{
			name: "invalid JSON",
			envVars: map[string]string{
				"DATAHUB_URL":                "https://datahub.example.com",
				"DATAHUB_TOKEN":              "test-token",
				"DATAHUB_ADDITIONAL_SERVERS": `{invalid json}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			for k := range tt.envVars {
				t.Setenv(k, "")
			}
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := FromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("FromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got := cfg.ConnectionCount(); got != tt.wantCount {
				t.Errorf("ConnectionCount() = %d, want %d", got, tt.wantCount)
			}
			if tt.wantDefault != "" && cfg.Default != tt.wantDefault {
				t.Errorf("Default = %q, want %q", cfg.Default, tt.wantDefault)
			}
		})
	}
}

func TestConfig_ClientConfig(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:             "https://prod.datahub.example.com",
			Token:           "prod-token",
			Timeout:         30 * time.Second,
			RetryMax:        3,
			DefaultLimit:    10,
			MaxLimit:        100,
			MaxLineageDepth: 5,
		},
		Connections: map[string]ConnectionConfig{
			"staging": {
				URL:          "https://staging.datahub.example.com",
				DefaultLimit: 20, // overridden
			},
			"dev": {
				URL:     "https://dev.datahub.example.com",
				Token:   "dev-token", // overridden
				Timeout: 60,          // overridden
			},
		},
	}

	tests := []struct {
		name         string
		connName     string
		wantURL      string
		wantToken    string
		wantTimeout  time.Duration
		wantDefLimit int
		wantErr      bool
	}{
		{
			name:         "default connection",
			connName:     "default",
			wantURL:      "https://prod.datahub.example.com",
			wantToken:    "prod-token",
			wantTimeout:  30 * time.Second,
			wantDefLimit: 10,
		},
		{
			name:         "empty name returns default",
			connName:     "",
			wantURL:      "https://prod.datahub.example.com",
			wantToken:    "prod-token",
			wantTimeout:  30 * time.Second,
			wantDefLimit: 10,
		},
		{
			name:         "staging inherits token",
			connName:     "staging",
			wantURL:      "https://staging.datahub.example.com",
			wantToken:    "prod-token", // inherited
			wantTimeout:  30 * time.Second,
			wantDefLimit: 20, // overridden
		},
		{
			name:         "dev with overrides",
			connName:     "dev",
			wantURL:      "https://dev.datahub.example.com",
			wantToken:    "dev-token",
			wantTimeout:  60 * time.Second,
			wantDefLimit: 10, // inherited
		},
		{
			name:     "unknown connection",
			connName: "unknown",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.ClientConfig(tt.connName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientConfig(%q) error = %v, wantErr %v", tt.connName, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}
			if got.Token != tt.wantToken {
				t.Errorf("Token = %q, want %q", got.Token, tt.wantToken)
			}
			if got.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", got.Timeout, tt.wantTimeout)
			}
			if got.DefaultLimit != tt.wantDefLimit {
				t.Errorf("DefaultLimit = %d, want %d", got.DefaultLimit, tt.wantDefLimit)
			}
		})
	}
}

func TestConfig_ConnectionNames(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{URL: "https://localhost"},
		Connections: map[string]ConnectionConfig{
			"staging": {URL: "https://staging.example.com"},
			"dev":     {URL: "https://dev.example.com"},
		},
	}

	names := cfg.ConnectionNames()
	if len(names) != 3 {
		t.Errorf("ConnectionNames() returned %d names, want 3", len(names))
	}
	if names[0] != "default" {
		t.Errorf("First name = %q, want %q", names[0], "default")
	}
}

func TestConfig_ConnectionInfos(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]ConnectionConfig{
			"staging": {URL: "https://staging.datahub.example.com"},
		},
	}

	infos := cfg.ConnectionInfos()
	if len(infos) != 2 {
		t.Errorf("ConnectionInfos() returned %d infos, want 2", len(infos))
	}

	// Check default connection
	var defaultInfo *ConnectionInfo
	for i := range infos {
		if infos[i].Name == "default" {
			defaultInfo = &infos[i]
			break
		}
	}
	if defaultInfo == nil {
		t.Fatal("default connection not found in infos")
	}
	if !defaultInfo.IsDefault {
		t.Error("default connection should have IsDefault = true")
	}
	if defaultInfo.URL != "https://prod.datahub.example.com" {
		t.Errorf("default URL = %q, want %q", defaultInfo.URL, "https://prod.datahub.example.com")
	}
}

func TestManager_HasConnection(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{URL: "https://localhost", Token: "token"},
		Connections: map[string]ConnectionConfig{
			"staging": {URL: "https://staging.example.com"},
		},
	}
	mgr := NewManager(cfg)

	tests := []struct {
		name string
		want bool
	}{
		{"", true}, // empty = default
		{"default", true},
		{"staging", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mgr.HasConnection(tt.name); got != tt.want {
				t.Errorf("HasConnection(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestManager_Connections(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{URL: "https://localhost", Token: "token"},
		Connections: map[string]ConnectionConfig{
			"staging": {URL: "https://staging.example.com"},
			"dev":     {URL: "https://dev.example.com"},
		},
	}
	mgr := NewManager(cfg)

	conns := mgr.Connections()
	if len(conns) != 3 {
		t.Errorf("Connections() returned %d, want 3", len(conns))
	}
}

func TestManager_ConnectionCount(t *testing.T) {
	cfg := Config{
		Default:     "default",
		Primary:     client.Config{URL: "https://localhost", Token: "token"},
		Connections: map[string]ConnectionConfig{},
	}
	mgr := NewManager(cfg)

	if got := mgr.ConnectionCount(); got != 1 {
		t.Errorf("ConnectionCount() = %d, want 1", got)
	}

	cfg.Connections["staging"] = ConnectionConfig{URL: "https://staging.example.com"}
	mgr = NewManager(cfg)
	if got := mgr.ConnectionCount(); got != 2 {
		t.Errorf("ConnectionCount() = %d, want 2", got)
	}
}

func TestManager_Client_UnknownConnection(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{URL: "https://localhost", Token: "token"},
	}
	mgr := NewManager(cfg)

	_, err := mgr.Client("unknown")
	if err == nil {
		t.Error("Client(\"unknown\") should return error")
	}
}

func TestNewManagerFromEnv(t *testing.T) {
	// Set up environment for valid config
	t.Setenv("DATAHUB_URL", "https://datahub.example.com")
	t.Setenv("DATAHUB_TOKEN", "test-token")
	t.Setenv("DATAHUB_ADDITIONAL_SERVERS", "")

	mgr, err := NewManagerFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
	if mgr.ConnectionCount() != 1 {
		t.Errorf("expected 1 connection, got %d", mgr.ConnectionCount())
	}
}

func TestNewManagerFromEnv_WithAdditionalServers(t *testing.T) {
	t.Setenv("DATAHUB_URL", "https://prod.datahub.example.com")
	t.Setenv("DATAHUB_TOKEN", "prod-token")
	t.Setenv("DATAHUB_ADDITIONAL_SERVERS", `{"staging": {"url": "https://staging.datahub.example.com", "token": "staging-token"}}`)

	mgr, err := NewManagerFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mgr.ConnectionCount() != 2 {
		t.Errorf("expected 2 connections, got %d", mgr.ConnectionCount())
	}
}

func TestNewManagerFromEnv_InvalidJSON(t *testing.T) {
	t.Setenv("DATAHUB_URL", "https://datahub.example.com")
	t.Setenv("DATAHUB_TOKEN", "test-token")
	t.Setenv("DATAHUB_ADDITIONAL_SERVERS", `{invalid}`)

	_, err := NewManagerFromEnv()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestManager_DefaultClient(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "test-token",
		},
	}
	mgr := NewManager(cfg)

	// DefaultClient should return the same as Client("default")
	c1, err := mgr.DefaultClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c2, err := mgr.Client("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c1 != c2 {
		t.Error("DefaultClient should return same client as Client(\"default\")")
	}

	// Clean up
	if err := mgr.Close(); err != nil {
		t.Errorf("unexpected error closing: %v", err)
	}
}

func TestManager_ConnectionInfos(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
		Connections: map[string]ConnectionConfig{
			"staging": {URL: "https://staging.datahub.example.com"},
		},
	}
	mgr := NewManager(cfg)

	infos := mgr.ConnectionInfos()
	if len(infos) != 2 {
		t.Errorf("expected 2 connection infos, got %d", len(infos))
	}

	// Verify default connection info
	var defaultInfo *ConnectionInfo
	for i := range infos {
		if infos[i].Name == "default" {
			defaultInfo = &infos[i]
			break
		}
	}
	if defaultInfo == nil {
		t.Fatal("default connection not found")
	}
	if !defaultInfo.IsDefault {
		t.Error("default connection should have IsDefault=true")
	}
	if defaultInfo.URL != "https://prod.datahub.example.com" {
		t.Errorf("expected URL 'https://prod.datahub.example.com', got %q", defaultInfo.URL)
	}
}

func TestManager_Config(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "test-token",
		},
		Connections: map[string]ConnectionConfig{
			"staging": {URL: "https://staging.datahub.example.com"},
		},
	}
	mgr := NewManager(cfg)

	returnedCfg := mgr.Config()
	if returnedCfg.Default != cfg.Default {
		t.Errorf("expected Default %q, got %q", cfg.Default, returnedCfg.Default)
	}
	if returnedCfg.Primary.URL != cfg.Primary.URL {
		t.Errorf("expected Primary.URL %q, got %q", cfg.Primary.URL, returnedCfg.Primary.URL)
	}
	if len(returnedCfg.Connections) != len(cfg.Connections) {
		t.Errorf("expected %d connections, got %d", len(cfg.Connections), len(returnedCfg.Connections))
	}
}

func TestManager_Close(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "test-token",
		},
	}
	mgr := NewManager(cfg)

	// Create a client
	_, err := mgr.Client("default")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	// Close should work
	err = mgr.Close()
	if err != nil {
		t.Errorf("unexpected error closing: %v", err)
	}

	// After close, client cache should be empty (new client created on next access)
	// This is hard to test directly, but we can verify Close doesn't error on empty cache
	err = mgr.Close()
	if err != nil {
		t.Errorf("unexpected error closing empty manager: %v", err)
	}
}

func TestManager_Close_NoClients(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "test-token",
		},
	}
	mgr := NewManager(cfg)

	// Close without any clients created
	err := mgr.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSingleClientManager(t *testing.T) {
	// Create a mock config (we won't actually connect)
	cfg := client.Config{
		URL:   "https://test.datahub.example.com",
		Token: "test-token",
	}

	// Use nil client to test manager structure without actual connection
	mgr := SingleClientManager(nil, cfg)

	if mgr.ConnectionCount() != 1 {
		t.Errorf("expected 1 connection, got %d", mgr.ConnectionCount())
	}

	if !mgr.HasConnection("datahub") {
		t.Error("expected HasConnection(\"datahub\") to return true")
	}

	if mgr.HasConnection("staging") {
		t.Error("expected HasConnection(\"staging\") to return false")
	}

	conns := mgr.Connections()
	if len(conns) != 1 {
		t.Errorf("expected 1 connection name, got %d", len(conns))
	}
	if conns[0] != "datahub" {
		t.Errorf("expected connection name 'datahub', got %q", conns[0])
	}

	infos := mgr.ConnectionInfos()
	if len(infos) != 1 {
		t.Errorf("expected 1 connection info, got %d", len(infos))
	}

	// Config should be accessible
	returnedCfg := mgr.Config()
	if returnedCfg.Primary.URL != cfg.URL {
		t.Errorf("expected URL %q, got %q", cfg.URL, returnedCfg.Primary.URL)
	}
}

func TestManager_Client_ConcurrentAccess(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "test-token",
		},
	}
	mgr := NewManager(cfg)
	defer func() {
		if err := mgr.Close(); err != nil {
			t.Errorf("failed to close manager: %v", err)
		}
	}()

	// Concurrent access should be safe
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			// Intentionally ignore error - testing concurrent access safety
			if _, err := mgr.Client("default"); err != nil {
				// Error is expected in some cases, just testing for races
				_ = err
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestManager_Client_InheritsFromPrimary(t *testing.T) {
	cfg := Config{
		Default: "default",
		Primary: client.Config{
			URL:   "https://datahub.example.com",
			Token: "test-token",
		},
		Connections: map[string]ConnectionConfig{
			"staging": {
				// URL and Token empty - will inherit from primary
			},
		},
	}
	mgr := NewManager(cfg)
	defer func() {
		if err := mgr.Close(); err != nil {
			t.Errorf("failed to close manager: %v", err)
		}
	}()

	// Should succeed because URL and Token inherit from primary
	_, err := mgr.Client("staging")
	if err != nil {
		t.Errorf("unexpected error: %v - staging should inherit from primary", err)
	}
}
