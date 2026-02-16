package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/integration"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
)

// Toolkit provides MCP tools for DataHub operations.
// It's designed to be composable - you can add its tools to any MCP server.
type Toolkit struct {
	client  DataHubClient        // Single client mode (for backwards compatibility)
	manager *multiserver.Manager // Multi-server mode (optional)
	config  Config
	logger  client.Logger

	// Extensibility hooks (all optional, zero-value = no overhead)
	middlewares     []ToolMiddleware
	toolMiddlewares map[ToolName][]ToolMiddleware

	// Integration interfaces (all optional, set via With* options)
	urnResolver      integration.URNResolver
	accessFilter     integration.AccessFilter
	auditLogger      integration.AuditLogger
	metadataEnricher integration.MetadataEnricher
	getUserID        func(context.Context) string

	// Query execution context provider (optional)
	queryProvider integration.QueryProvider

	// Pre-built integration middleware (built after options applied)
	integrationMiddleware []ToolMiddleware

	// Description overrides (toolkit-level, set via WithDescriptions)
	descriptions map[ToolName]string

	// Annotation overrides (toolkit-level, set via WithAnnotations)
	annotations map[ToolName]*mcp.ToolAnnotations

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
	// Initialize logger
	logger := cfg.Logger
	if logger == nil {
		if cfg.Debug {
			logger = client.NewStdLogger(true)
		} else {
			logger = client.NopLogger{}
		}
	}

	return &Toolkit{
		config:          cfg,
		logger:          logger,
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		descriptions:    make(map[ToolName]string),
		annotations:     make(map[ToolName]*mcp.ToolAnnotations),
		registeredTools: make(map[ToolName]bool),
	}
}

// applyToolkitOptions applies options to the toolkit.
func applyToolkitOptions(t *Toolkit, opts []ToolkitOption) {
	for _, opt := range opts {
		opt(t)
	}
	t.buildIntegrationMiddleware()
}

// buildIntegrationMiddleware builds middleware adapters from integration interfaces.
func (t *Toolkit) buildIntegrationMiddleware() {
	var mws []ToolMiddleware

	// Order matters: resolve URN first, then check access
	if t.urnResolver != nil {
		mws = append(mws, NewURNResolverMiddleware(t.urnResolver))
	}
	if t.accessFilter != nil {
		mws = append(mws, NewAccessFilterMiddleware(t.accessFilter))
	}
	// Enrichment happens after handler, audit logs last
	if t.metadataEnricher != nil {
		mws = append(mws, NewMetadataEnricherMiddleware(t.metadataEnricher))
	}
	if t.auditLogger != nil {
		mws = append(mws, NewAuditLoggerMiddleware(t.auditLogger, t.getUserID))
	}

	t.integrationMiddleware = mws
}

// RegisterAll adds all DataHub tools to the given MCP server.
// If WriteEnabled is true, also registers write tools.
func (t *Toolkit) RegisterAll(server *mcp.Server) {
	t.Register(server, AllTools()...)
	if t.isWriteEnabled() {
		t.Register(server, WriteTools()...)
	}
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

// toolRegistrar is a function that registers a single tool on a server.
type toolRegistrar func(server *mcp.Server, cfg *toolConfig)

// toolRegistry returns the mapping of tool names to their registration functions.
func (t *Toolkit) toolRegistry() map[ToolName]toolRegistrar {
	return map[ToolName]toolRegistrar{
		// Read tools
		ToolSearch:           t.registerSearchTool,
		ToolGetEntity:        t.registerGetEntityTool,
		ToolGetSchema:        t.registerGetSchemaTool,
		ToolGetLineage:       t.registerGetLineageTool,
		ToolGetColumnLineage: t.registerGetColumnLineageTool,
		ToolGetQueries:       t.registerGetQueriesTool,
		ToolGetGlossaryTerm:  t.registerGetGlossaryTermTool,
		ToolListTags:         t.registerListTagsTool,
		ToolListDomains:      t.registerListDomainsTool,
		ToolListDataProducts: t.registerListDataProductsTool,
		ToolGetDataProduct:   t.registerGetDataProductTool,
		ToolListConnections:  t.registerListConnectionsTool,
		// Write tools
		ToolUpdateDescription:  t.registerUpdateDescriptionTool,
		ToolAddTag:             t.registerAddTagTool,
		ToolRemoveTag:          t.registerRemoveTagTool,
		ToolAddGlossaryTerm:    t.registerAddGlossaryTermTool,
		ToolRemoveGlossaryTerm: t.registerRemoveGlossaryTermTool,
		ToolAddLink:            t.registerAddLinkTool,
		ToolRemoveLink:         t.registerRemoveLinkTool,
	}
}

// registerTool is the internal registration method.
func (t *Toolkit) registerTool(server *mcp.Server, name ToolName, cfg *toolConfig) {
	if t.registeredTools[name] {
		return // Already registered
	}

	if register, ok := t.toolRegistry()[name]; ok {
		register(server, cfg)
		t.registeredTools[name] = true
	}
}

// wrapHandler wraps a handler with middleware support.
func (t *Toolkit) wrapHandler(
	name ToolName,
	handler func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error),
	cfg *toolConfig,
) func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
	// Collect all applicable middlewares
	// Integration middleware runs first (URN resolution, access control)
	var allMiddlewares []ToolMiddleware
	allMiddlewares = append(allMiddlewares, t.integrationMiddleware...)
	allMiddlewares = append(allMiddlewares, t.middlewares...)
	allMiddlewares = append(allMiddlewares, t.toolMiddlewares[name]...)
	if cfg != nil {
		allMiddlewares = append(allMiddlewares, cfg.middlewares...)
	}

	// If no middleware configured and no debug logging, return handler unchanged
	if len(allMiddlewares) == 0 && !t.config.Debug {
		return handler
	}

	return func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		tc := NewToolContext(name, input)
		start := time.Now()

		t.log().Debug("tool invoked",
			"tool", string(name),
			"input_type", fmt.Sprintf("%T", input),
			"middleware_count", len(allMiddlewares))

		// Run Before hooks
		var err error
		for _, m := range allMiddlewares {
			ctx, err = m.Before(ctx, tc)
			if err != nil {
				t.log().Error("middleware before hook failed",
					"tool", string(name),
					"error", err.Error())
				return ErrorResult(fmt.Sprintf("middleware error: %v", err)), nil, nil
			}
		}

		// Execute handler
		result, extra, handlerErr := handler(ctx, req, input)

		// Log handler result
		switch {
		case handlerErr != nil:
			t.log().Error("tool handler error",
				"tool", string(name),
				"error", handlerErr.Error(),
				"duration_ms", time.Since(start).Milliseconds())
		case result != nil && result.IsError:
			t.log().Debug("tool returned error result",
				"tool", string(name),
				"duration_ms", time.Since(start).Milliseconds())
		default:
			t.log().Debug("tool handler completed",
				"tool", string(name),
				"duration_ms", time.Since(start).Milliseconds())
		}

		// Run After hooks (reverse order)
		for i := len(allMiddlewares) - 1; i >= 0; i-- {
			result, err = allMiddlewares[i].After(ctx, tc, result, handlerErr)
			if err != nil {
				t.log().Error("middleware after hook failed",
					"tool", string(name),
					"error", err.Error())
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
	if len(t.integrationMiddleware) > 0 {
		return true
	}
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

// QueryProvider returns the configured query provider, or nil if not configured.
func (t *Toolkit) QueryProvider() integration.QueryProvider {
	return t.queryProvider
}

// HasQueryProvider returns true if a query provider is configured.
func (t *Toolkit) HasQueryProvider() bool {
	return t.queryProvider != nil
}

// isWriteEnabled returns true if write operations are enabled in the toolkit config.
func (t *Toolkit) isWriteEnabled() bool {
	return t.config.WriteEnabled
}

// getWriteClient returns the DataHub client for write operations.
// Returns ErrWriteDisabled if write operations are not enabled.
func (t *Toolkit) getWriteClient(connection string) (DataHubClient, error) {
	if !t.isWriteEnabled() {
		return nil, client.ErrWriteDisabled
	}
	return t.getClient(connection)
}

// getClient returns the DataHub client for the given connection name.
// If connection is empty, returns the default client.
// In single-client mode, always returns the single client.
func (t *Toolkit) getClient(connection string) (DataHubClient, error) {
	// Multi-server mode
	if t.manager != nil {
		connName := connection
		if connName == "" {
			connName = "(default)"
		}
		t.log().Debug("selecting connection", "connection", connName)
		c, err := t.manager.Client(connection)
		if err != nil {
			t.log().Error("connection selection failed",
				"connection", connName,
				"error", err.Error())
		}
		return c, err
	}

	// Single-client mode - ignore connection parameter
	if t.client == nil {
		t.log().Error("no client configured")
		return nil, fmt.Errorf("no client configured")
	}
	return t.client, nil
}

// log returns the logger, defaulting to NopLogger if nil.
func (t *Toolkit) log() client.Logger {
	if t.logger == nil {
		return client.NopLogger{}
	}
	return t.logger
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
