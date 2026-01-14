package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolMiddleware intercepts tool execution for cross-cutting concerns.
type ToolMiddleware interface {
	// Before is called before the tool handler executes.
	// Return a modified context or an error to abort execution.
	Before(ctx context.Context, tc *ToolContext) (context.Context, error)

	// After is called after the tool handler executes.
	// Can modify the result or handle errors.
	After(ctx context.Context, tc *ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error)
}

// BeforeFunc is a convenience type for creating Before-only middleware.
type BeforeFunc func(ctx context.Context, tc *ToolContext) (context.Context, error)

// Before implements ToolMiddleware.
func (f BeforeFunc) Before(ctx context.Context, tc *ToolContext) (context.Context, error) {
	return f(ctx, tc)
}

// After implements ToolMiddleware (no-op).
func (f BeforeFunc) After(_ context.Context, _ *ToolContext, result *mcp.CallToolResult, _ error) (*mcp.CallToolResult, error) {
	return result, nil
}

// AfterFunc is a convenience type for creating After-only middleware.
type AfterFunc func(ctx context.Context, tc *ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error)

// Before implements ToolMiddleware (no-op).
func (f AfterFunc) Before(ctx context.Context, _ *ToolContext) (context.Context, error) {
	return ctx, nil
}

// After implements ToolMiddleware.
func (f AfterFunc) After(ctx context.Context, tc *ToolContext, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
	return f(ctx, tc, result, err)
}
