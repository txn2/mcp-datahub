package multiserver

import (
	"fmt"
	"sync"

	"github.com/txn2/mcp-datahub/pkg/client"
)

// Manager manages connections to multiple DataHub servers.
// It lazily creates client connections on first use.
type Manager struct {
	config  Config
	clients map[string]*client.Client
	mu      sync.RWMutex
}

// NewManager creates a new connection manager with the given configuration.
// Clients are created lazily on first access, not at construction time.
func NewManager(cfg Config) *Manager {
	return &Manager{
		config:  cfg,
		clients: make(map[string]*client.Client),
	}
}

// NewManagerFromEnv creates a Manager using configuration from environment variables.
func NewManagerFromEnv() (*Manager, error) {
	cfg, err := FromEnv()
	if err != nil {
		return nil, err
	}
	return NewManager(cfg), nil
}

// Client returns the DataHub client for the named connection.
// If name is empty, returns the primary connection's client.
// Clients are created lazily and cached for reuse.
func (m *Manager) Client(name string) (*client.Client, error) {
	// Normalize empty to default
	if name == "" {
		name = m.config.Default
	}

	// Check cache first (read lock)
	m.mu.RLock()
	if c, ok := m.clients[name]; ok {
		m.mu.RUnlock()
		return c, nil
	}
	m.mu.RUnlock()

	// Need to create client (write lock)
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if c, ok := m.clients[name]; ok {
		return c, nil
	}

	// Get client config
	cfg, err := m.config.ClientConfig(name)
	if err != nil {
		return nil, err
	}

	// Validate config
	if validateErr := cfg.Validate(); validateErr != nil {
		return nil, fmt.Errorf("invalid config for connection %q: %w", name, validateErr)
	}

	// Create client
	c, err := client.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating client for connection %q: %w", name, err)
	}

	m.clients[name] = c
	return c, nil
}

// DefaultClient returns the default (primary) connection's client.
func (m *Manager) DefaultClient() (*client.Client, error) {
	return m.Client(m.config.Default)
}

// Connections returns the names of all configured connections.
func (m *Manager) Connections() []string {
	return m.config.ConnectionNames()
}

// ConnectionInfos returns information about all configured connections.
func (m *Manager) ConnectionInfos() []ConnectionInfo {
	return m.config.ConnectionInfos()
}

// ConnectionCount returns the number of configured connections.
func (m *Manager) ConnectionCount() int {
	return m.config.ConnectionCount()
}

// HasConnection returns true if the named connection exists.
func (m *Manager) HasConnection(name string) bool {
	if name == "" || name == m.config.Default {
		return true
	}
	_, ok := m.config.Connections[name]
	return ok
}

// Config returns the manager's configuration.
func (m *Manager) Config() Config {
	return m.config
}

// Close closes all open client connections.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for name, c := range m.clients {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("closing connection %q: %w", name, err)
		}
	}
	m.clients = make(map[string]*client.Client)
	return firstErr
}

// SingleClientManager creates a Manager with only a default connection.
// This is useful for backwards compatibility with code that uses a single client.
func SingleClientManager(c *client.Client, cfg client.Config) *Manager {
	return &Manager{
		config: Config{
			Default:     DefaultConnectionName,
			Primary:     cfg,
			Connections: make(map[string]ConnectionConfig),
		},
		clients: map[string]*client.Client{
			DefaultConnectionName: c,
		},
	}
}
