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

func TestWithDescription(t *testing.T) {
	cfg := &toolConfig{}
	opt := WithDescription("custom description")
	opt(cfg)

	if cfg.description == nil {
		t.Fatal("WithDescription() should set description pointer")
	}
	if *cfg.description != "custom description" {
		t.Errorf("WithDescription() description = %q, want %q", *cfg.description, "custom description")
	}
}

func TestWithDescriptions(t *testing.T) {
	toolkit := &Toolkit{
		descriptions:    make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	descs := map[ToolName]string{
		ToolSearch:    "custom search",
		ToolGetEntity: "custom entity",
	}

	opt := WithDescriptions(descs)
	opt(toolkit)

	if toolkit.descriptions[ToolSearch] != "custom search" {
		t.Errorf("WithDescriptions() search = %q, want %q", toolkit.descriptions[ToolSearch], "custom search")
	}
	if toolkit.descriptions[ToolGetEntity] != "custom entity" {
		t.Errorf("WithDescriptions() entity = %q, want %q", toolkit.descriptions[ToolGetEntity], "custom entity")
	}
}

func TestWithDescriptions_Merge(t *testing.T) {
	toolkit := &Toolkit{
		descriptions:    make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	// First batch
	WithDescriptions(map[ToolName]string{
		ToolSearch:    "first search",
		ToolGetEntity: "first entity",
	})(toolkit)

	// Second batch should merge (overwrite search, keep entity)
	WithDescriptions(map[ToolName]string{
		ToolSearch:    "second search",
		ToolGetSchema: "second schema",
	})(toolkit)

	if toolkit.descriptions[ToolSearch] != "second search" {
		t.Errorf("WithDescriptions() merge: search = %q, want %q", toolkit.descriptions[ToolSearch], "second search")
	}
	if toolkit.descriptions[ToolGetEntity] != "first entity" {
		t.Errorf("WithDescriptions() merge: entity = %q, want %q", toolkit.descriptions[ToolGetEntity], "first entity")
	}
	if toolkit.descriptions[ToolGetSchema] != "second schema" {
		t.Errorf("WithDescriptions() merge: schema = %q, want %q", toolkit.descriptions[ToolGetSchema], "second schema")
	}
}

func TestWithTitle(t *testing.T) {
	cfg := &toolConfig{}
	opt := WithTitle("My Custom Title")
	opt(cfg)

	if cfg.title == nil {
		t.Fatal("WithTitle() should set title pointer")
	}
	if *cfg.title != "My Custom Title" {
		t.Errorf("WithTitle() title = %q, want %q", *cfg.title, "My Custom Title")
	}
}

func TestWithTitles(t *testing.T) {
	toolkit := &Toolkit{
		titles:          make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	titles := map[ToolName]string{
		ToolSearch:    "Custom Search",
		ToolGetEntity: "Custom Entity",
	}

	opt := WithTitles(titles)
	opt(toolkit)

	if toolkit.titles[ToolSearch] != "Custom Search" {
		t.Errorf("WithTitles() search = %q, want %q", toolkit.titles[ToolSearch], "Custom Search")
	}
	if toolkit.titles[ToolGetEntity] != "Custom Entity" {
		t.Errorf("WithTitles() entity = %q, want %q", toolkit.titles[ToolGetEntity], "Custom Entity")
	}
}

func TestWithTitles_Merge(t *testing.T) {
	toolkit := &Toolkit{
		titles:          make(map[ToolName]string),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	WithTitles(map[ToolName]string{
		ToolSearch:    "first search",
		ToolGetEntity: "first entity",
	})(toolkit)

	WithTitles(map[ToolName]string{
		ToolSearch:    "second search",
		ToolGetSchema: "second schema",
	})(toolkit)

	if toolkit.titles[ToolSearch] != "second search" {
		t.Errorf("WithTitles() merge: search = %q, want %q", toolkit.titles[ToolSearch], "second search")
	}
	if toolkit.titles[ToolGetEntity] != "first entity" {
		t.Errorf("WithTitles() merge: entity = %q, want %q", toolkit.titles[ToolGetEntity], "first entity")
	}
	if toolkit.titles[ToolGetSchema] != "second schema" {
		t.Errorf("WithTitles() merge: schema = %q, want %q", toolkit.titles[ToolGetSchema], "second schema")
	}
}

func TestWithOutputSchema(t *testing.T) {
	customSchema := map[string]any{"type": "object"}
	cfg := &toolConfig{}
	opt := WithOutputSchema(customSchema)
	opt(cfg)

	if cfg.outputSchema == nil {
		t.Fatal("WithOutputSchema() should set outputSchema")
	}
	got, ok := cfg.outputSchema.(map[string]any)
	if !ok {
		t.Fatal("outputSchema type assertion failed")
	}
	if got["type"] != "object" {
		t.Errorf("WithOutputSchema() type = %v, want %q", got["type"], "object")
	}
}

func TestWithOutputSchemas(t *testing.T) {
	toolkit := &Toolkit{
		outputSchemas:   make(map[ToolName]any),
		toolMiddlewares: make(map[ToolName][]ToolMiddleware),
		registeredTools: make(map[ToolName]bool),
	}

	schemas := map[ToolName]any{
		ToolSearch:    map[string]any{"type": "object", "title": "search"},
		ToolGetEntity: map[string]any{"type": "object", "title": "entity"},
	}

	opt := WithOutputSchemas(schemas)
	opt(toolkit)

	if toolkit.outputSchemas[ToolSearch] == nil {
		t.Error("WithOutputSchemas() should set ToolSearch schema")
	}
	if toolkit.outputSchemas[ToolGetEntity] == nil {
		t.Error("WithOutputSchemas() should set ToolGetEntity schema")
	}
	if toolkit.outputSchemas[ToolGetSchema] != nil {
		t.Error("WithOutputSchemas() should not set ToolGetSchema schema")
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
