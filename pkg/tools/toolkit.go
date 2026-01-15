package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/multiserver"
)

// Toolkit provides MCP tools for DataHub operations.
// It's designed to be composable - you can add its tools to any MCP server.
type Toolkit struct {
	client  DataHubClient        // Single client mode (for backwards compatibility)
	manager *multiserver.Manager // Multi-server mode (optional)
	config  Config

	// Extensibility hooks (all optional, zero-value = no overhead)
	middlewares     []ToolMiddleware
	toolMiddlewares map[ToolName][]ToolMiddleware

	// Internal tracking
	registeredTools map[ToolName]bool
}

// NewToolkit creates a new DataHub toolkit.
// Accepts optional ToolkitOption arguments for middleware, etc.
// Maintains backwards compatibility - existing code works unchanged.
func NewToolkit(c DataHubClient, cfg Config, opts ...ToolkitOption) *Toolkit {
	t := newBaseToolkit(normalizeConfig(cfg))
	t.client = c
	applyToolkitOptions(t, opts)
	return t
}

// NewToolkitWithManager creates a Toolkit with multi-server support.
// Use this when you need to connect to multiple DataHub servers.
func NewToolkitWithManager(mgr *multiserver.Manager, cfg Config, opts ...ToolkitOption) *Toolkit {
	t := newBaseToolkit(normalizeConfig(cfg))
	t.manager = mgr
	applyToolkitOptions(t, opts)
	return t
}

// newBaseToolkit creates a toolkit with common fields initialized.
func newBaseToolkit(cfg Config) *Toolkit {
	return &Toolkit{
		config:          cfg,
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}
}

// applyToolkitOptions applies options to the toolkit.
func applyToolkitOptions(t *Toolkit, opts []ToolkitOption) {
	for _, opt := range opts {
		opt(t)
	}
}

// RegisterAll adds all DataHub tools to the given MCP server.
func (t *Toolkit) RegisterAll(server *mcp.Server) {
	t.Register(server, AllTools()...)
}

// Register adds specific tools to the server.
func (t *Toolkit) Register(server *mcp.Server, names ...ToolName) {
	for _, name := range names {
		t.registerTool(server, name, nil)
	}
}

// RegisterWith adds a tool with additional per-registration options.
func (t *Toolkit) RegisterWith(server *mcp.Server, name ToolName, opts ...ToolOption) {
	cfg := &toolConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	t.registerTool(server, name, cfg)
}

// registerTool is the internal registration method.
func (t *Toolkit) registerTool(server *mcp.Server, name ToolName, cfg *toolConfig) {
	if t.registeredTools[name] {
		return // Already registered
	}

	switch name {
	case ToolSearch:
		t.registerSearchTool(server, cfg)
	case ToolGetEntity:
		t.registerGetEntityTool(server, cfg)
	case ToolGetSchema:
		t.registerGetSchemaTool(server, cfg)
	case ToolGetLineage:
		t.registerGetLineageTool(server, cfg)
	case ToolGetQueries:
		t.registerGetQueriesTool(server, cfg)
	case ToolGetGlossaryTerm:
		t.registerGetGlossaryTermTool(server, cfg)
	case ToolListTags:
		t.registerListTagsTool(server, cfg)
	case ToolListDomains:
		t.registerListDomainsTool(server, cfg)
	case ToolListDataProducts:
		t.registerListDataProductsTool(server, cfg)
	case ToolGetDataProduct:
		t.registerGetDataProductTool(server, cfg)
	case ToolListConnections:
		t.registerListConnectionsTool(server, cfg)
	}

	t.registeredTools[name] = true
}

// wrapHandler wraps a handler with middleware support.
func (t *Toolkit) wrapHandler(
	name ToolName,
	handler func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error),
	cfg *toolConfig,
) func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
	// Collect all applicable middlewares
	var allMiddlewares []ToolMiddleware
	allMiddlewares = append(allMiddlewares, t.middlewares...)
	allMiddlewares = append(allMiddlewares, t.toolMiddlewares[name]...)
	if cfg != nil {
		allMiddlewares = append(allMiddlewares, cfg.middlewares...)
	}

	// If no middleware configured, return handler unchanged
	if len(allMiddlewares) == 0 {
		return handler
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		tc := NewToolContext(name, input)

		// Run Before hooks
		var err error
		for _, m := range allMiddlewares {
			ctx, err = m.Before(ctx, tc)
			if err != nil {
				return ErrorResult(fmt.Sprintf("middleware error: %v", err)), nil, nil
			}
		}

		// Execute handler
		result, extra, handlerErr := handler(ctx, req, input)

		// Run After hooks (reverse order)
		for i := len(allMiddlewares) - 1; i >= 0; i-- {
			result, err = allMiddlewares[i].After(ctx, tc, result, handlerErr)
			if err != nil {
				return ErrorResult(fmt.Sprintf("middleware error: %v", err)), nil, nil
			}
		}

		return result, extra, nil
	}
}

// Client returns the underlying DataHub client.
func (t *Toolkit) Client() DataHubClient {
	return t.client
}

// Config returns the toolkit configuration.
func (t *Toolkit) Config() Config {
	return t.config
}

// HasMiddleware returns true if any middleware is configured.
func (t *Toolkit) HasMiddleware() bool {
	if len(t.middlewares) > 0 {
		return true
	}
	for _, mws := range t.toolMiddlewares {
		if len(mws) > 0 {
			return true
		}
	}
	return false
}

// getClient returns the DataHub client for the given connection name.
// If connection is empty, returns the default client.
// In single-client mode, always returns the single client.
func (t *Toolkit) getClient(connection string) (DataHubClient, error) {
	// Multi-server mode
	if t.manager != nil {
		return t.manager.Client(connection)
	}

	// Single-client mode - ignore connection parameter
	if t.client == nil {
		return nil, fmt.Errorf("no client configured")
	}
	return t.client, nil
}

// HasManager returns true if multi-server mode is enabled.
func (t *Toolkit) HasManager() bool {
	return t.manager != nil
}

// Manager returns the connection manager, or nil if in single-client mode.
func (t *Toolkit) Manager() *multiserver.Manager {
	return t.manager
}

// ConnectionInfos returns information about all configured connections.
// Returns a single "default" connection in single-client mode.
func (t *Toolkit) ConnectionInfos() []multiserver.ConnectionInfo {
	if t.manager != nil {
		return t.manager.ConnectionInfos()
	}

	// Single-client mode - return default connection info
	return []multiserver.ConnectionInfo{
		{
			Name:      "default",
			URL:       "configured via single client",
			IsDefault: true,
		},
	}
}

// ConnectionCount returns the number of configured connections.
func (t *Toolkit) ConnectionCount() int {
	if t.manager != nil {
		return t.manager.ConnectionCount()
	}
	return 1
}
