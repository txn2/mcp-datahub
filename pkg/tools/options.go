package tools

// ToolkitOption configures a Toolkit.
type ToolkitOption func(*Toolkit)

// WithMiddleware adds global middleware to all tools.
func WithMiddleware(mw ToolMiddleware) ToolkitOption {
	return func(t *Toolkit) {
		t.middlewares = append(t.middlewares, mw)
	}
}

// WithToolMiddleware adds middleware to a specific tool.
func WithToolMiddleware(name ToolName, mw ToolMiddleware) ToolkitOption {
	return func(t *Toolkit) {
		t.toolMiddlewares[name] = append(t.toolMiddlewares[name], mw)
	}
}

// toolConfig holds per-registration configuration.
type toolConfig struct {
	middlewares []ToolMiddleware
}

// ToolOption configures a single tool registration.
type ToolOption func(*toolConfig)

// WithPerToolMiddleware adds middleware for a single tool registration.
func WithPerToolMiddleware(mw ToolMiddleware) ToolOption {
	return func(cfg *toolConfig) {
		cfg.middlewares = append(cfg.middlewares, mw)
	}
}
