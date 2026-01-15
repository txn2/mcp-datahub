// Package multiserver provides support for managing connections to multiple DataHub servers.
package multiserver

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/txn2/mcp-datahub/pkg/client"
)

// DefaultConnectionName is the fallback name for the primary connection.
const DefaultConnectionName = "datahub"

// ConnectionConfig defines configuration for a single DataHub connection.
// Fields that are empty/zero inherit from the primary connection.
type ConnectionConfig struct {
	// URL is the DataHub GMS URL (required for additional servers).
	URL string `json:"url"`

	// Token is the personal access token. Inherits from primary if empty.
	Token string `json:"token,omitempty"`

	// Timeout is the request timeout in seconds. Inherits from primary if zero.
	Timeout int `json:"timeout,omitempty"`

	// RetryMax is the maximum retry attempts. Inherits from primary if zero.
	RetryMax int `json:"retry_max,omitempty"`

	// DefaultLimit is the default search result limit. Inherits from primary if zero.
	DefaultLimit int `json:"default_limit,omitempty"`

	// MaxLimit is the maximum allowed limit. Inherits from primary if zero.
	MaxLimit int `json:"max_limit,omitempty"`

	// MaxLineageDepth is the maximum lineage traversal depth. Inherits from primary if zero.
	MaxLineageDepth int `json:"max_lineage_depth,omitempty"`
}

// Config holds configuration for multiple DataHub connections.
type Config struct {
	// Default is the display name of the primary connection.
	// This can be customized via DATAHUB_CONNECTION_NAME env var.
	// Defaults to "datahub" if not specified.
	Default string

	// Primary is the primary connection configuration (from DATAHUB_* env vars).
	Primary client.Config

	// Connections maps connection names to their configurations.
	// The primary connection is always available under the Default name.
	Connections map[string]ConnectionConfig
}

// FromEnv builds a multi-server configuration from environment variables.
//
// Primary server configuration comes from standard DATAHUB_* variables.
// Additional servers come from DATAHUB_ADDITIONAL_SERVERS as JSON:
//
//	{"staging": {"url": "https://staging.datahub.example.com", "token": "xxx"}}
//
// The primary connection name can be customized via DATAHUB_CONNECTION_NAME
// (defaults to "datahub").
//
// Additional servers inherit token, timeout, retry_max, etc. from the primary
// if not explicitly specified.
func FromEnv() (Config, error) {
	primary, err := client.FromEnv()
	if err != nil {
		return Config{}, fmt.Errorf("loading primary config: %w", err)
	}

	// Get connection name from env or use default
	connectionName := os.Getenv("DATAHUB_CONNECTION_NAME")
	if connectionName == "" {
		connectionName = DefaultConnectionName
	}

	cfg := Config{
		Default:     connectionName,
		Primary:     primary,
		Connections: make(map[string]ConnectionConfig),
	}

	// Parse additional servers from JSON env var
	additionalJSON := os.Getenv("DATAHUB_ADDITIONAL_SERVERS")
	if additionalJSON != "" {
		var additional map[string]ConnectionConfig
		if err := json.Unmarshal([]byte(additionalJSON), &additional); err != nil {
			return Config{}, fmt.Errorf("parsing DATAHUB_ADDITIONAL_SERVERS: %w", err)
		}
		cfg.Connections = additional
	}

	return cfg, nil
}

// ClientConfig returns a client.Config for the named connection.
// Returns the primary config if name is empty or matches the default connection name.
// Returns an error if the connection name is not found.
func (c Config) ClientConfig(name string) (client.Config, error) {
	// Empty or default connection name returns primary
	if name == "" || name == c.Default {
		return c.Primary, nil
	}

	// Look up additional connection
	conn, ok := c.Connections[name]
	if !ok {
		return client.Config{}, fmt.Errorf("unknown connection: %q (available: %v)", name, c.ConnectionNames())
	}

	// Build config by inheriting from primary
	cfg := c.Primary // Start with primary values

	// Override with connection-specific values
	if conn.URL != "" {
		cfg.URL = conn.URL
	}
	if conn.Token != "" {
		cfg.Token = conn.Token
	}
	if conn.Timeout > 0 {
		cfg.Timeout = time.Duration(conn.Timeout) * time.Second
	}
	if conn.RetryMax > 0 {
		cfg.RetryMax = conn.RetryMax
	}
	if conn.DefaultLimit > 0 {
		cfg.DefaultLimit = conn.DefaultLimit
	}
	if conn.MaxLimit > 0 {
		cfg.MaxLimit = conn.MaxLimit
	}
	if conn.MaxLineageDepth > 0 {
		cfg.MaxLineageDepth = conn.MaxLineageDepth
	}

	return cfg, nil
}

// ConnectionNames returns the names of all available connections.
// Always includes the primary connection name as the first entry.
func (c Config) ConnectionNames() []string {
	names := []string{c.Default}
	for name := range c.Connections {
		names = append(names, name)
	}
	return names
}

// ConnectionCount returns the total number of connections (including default).
func (c Config) ConnectionCount() int {
	return 1 + len(c.Connections)
}

// ConnectionInfo holds display information about a connection.
type ConnectionInfo struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	IsDefault bool   `json:"is_default"`
}

// ConnectionInfos returns information about all connections for display.
func (c Config) ConnectionInfos() []ConnectionInfo {
	infos := make([]ConnectionInfo, 0, c.ConnectionCount())

	// Add primary connection
	infos = append(infos, ConnectionInfo{
		Name:      c.Default,
		URL:       c.Primary.URL,
		IsDefault: true,
	})

	// Add additional connections
	for name := range c.Connections {
		cfg, err := c.ClientConfig(name)
		if err != nil {
			continue // Should not happen for known connections
		}
		infos = append(infos, ConnectionInfo{
			Name:      name,
			URL:       cfg.URL,
			IsDefault: false,
		})
	}

	return infos
}
