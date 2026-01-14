package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Toolkit provides MCP tools for DataHub operations.
// It's designed to be composable - you can add its tools to any MCP server.
type Toolkit struct {
	client DataHubClient
	config Config

	// Extensibility hooks (all optional, zero-value = no overhead)
	middlewares     []ToolMiddleware
	toolMiddlewares map[ToolName][]ToolMiddleware

	// Internal tracking
	registeredTools map[ToolName]bool
}

// NewToolkit creates a new DataHub toolkit.
func NewToolkit(c DataHubClient, cfg Config, opts ...ToolkitOption) *Toolkit {
	t := &Toolkit{
		client:          c,
		config:          normalizeConfig(cfg),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
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
