package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestWithMiddleware(t *testing.T) {
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})

	toolkit := &Toolkit{
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	opt := WithMiddleware(mw)
	opt(toolkit)

	if len(toolkit.middlewares) != 1 {
		t.Errorf("WithMiddleware() should add 1 middleware, got %d", len(toolkit.middlewares))
	}
}

func TestWithToolMiddleware(t *testing.T) {
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})

	toolkit := &Toolkit{
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	opt := WithToolMiddleware(ToolSearch, mw)
	opt(toolkit)

	if len(toolkit.toolMiddlewares[ToolSearch]) != 1 {
		t.Errorf("WithToolMiddleware() should add 1 middleware for ToolSearch, got %d", len(toolkit.toolMiddlewares[ToolSearch]))
	}
	if len(toolkit.toolMiddlewares[ToolGetEntity]) != 0 {
		t.Error("WithToolMiddleware() should not affect other tools")
	}
}

func TestWithPerToolMiddleware(t *testing.T) {
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})

	cfg := &toolConfig{}
	opt := WithPerToolMiddleware(mw)
	opt(cfg)

	if len(cfg.middlewares) != 1 {
		t.Errorf("WithPerToolMiddleware() should add 1 middleware, got %d", len(cfg.middlewares))
	}
}

func TestMultipleMiddleware(t *testing.T) {
	mw1 := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})
	mw2 := AfterFunc(func(_ context.Context, _ *ToolContext, result *mcp.CallToolResult, _ error) (*mcp.CallToolResult, error) {
		return result, nil
	})

	toolkit := &Toolkit{
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	WithMiddleware(mw1)(toolkit)
	WithMiddleware(mw2)(toolkit)

	if len(toolkit.middlewares) != 2 {
		t.Errorf("Multiple WithMiddleware() calls should add 2 middlewares, got %d", len(toolkit.middlewares))
	}
}

func TestMultipleToolMiddleware(t *testing.T) {
	mw := BeforeFunc(func(ctx context.Context, _ *ToolContext) (context.Context, error) {
		return ctx, nil
	})

	toolkit := &Toolkit{
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	WithToolMiddleware(ToolSearch, mw)(toolkit)
	WithToolMiddleware(ToolSearch, mw)(toolkit)
	WithToolMiddleware(ToolGetEntity, mw)(toolkit)

	if len(toolkit.toolMiddlewares[ToolSearch]) != 2 {
		t.Errorf("Multiple WithToolMiddleware() for same tool should add 2 middlewares, got %d", len(toolkit.toolMiddlewares[ToolSearch]))
	}
	if len(toolkit.toolMiddlewares[ToolGetEntity]) != 1 {
		t.Errorf("WithToolMiddleware() for different tool should add 1 middleware, got %d", len(toolkit.toolMiddlewares[ToolGetEntity]))
	}
}
