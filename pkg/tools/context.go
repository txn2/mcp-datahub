package tools

import "time"

// ToolContext provides context for middleware about the current tool execution.
type ToolContext struct {
	// ToolName is the name of the tool being executed.
	ToolName ToolName

	// Input is the parsed input for the tool.
	Input any

	// StartTime is when the tool execution started.
	StartTime time.Time

	// Extra allows middleware to pass data between Before and After.
	Extra map[string]any
}

// NewToolContext creates a new ToolContext.
func NewToolContext(name ToolName, input any) *ToolContext {
	return &ToolContext{
		ToolName:  name,
		Input:     input,
		StartTime: time.Now(),
		Extra:     make(map[string]any),
	}
}

// Duration returns the time elapsed since the tool started.
func (tc *ToolContext) Duration() time.Duration {
	return time.Since(tc.StartTime)
}

// Set stores a value in Extra.
func (tc *ToolContext) Set(key string, value any) {
	tc.Extra[key] = value
}

// Get retrieves a value from Extra.
func (tc *ToolContext) Get(key string) (any, bool) {
	v, ok := tc.Extra[key]
	return v, ok
}

// GetString retrieves a string value from Extra.
// Returns an empty string if the key is not found or the value is not a string.
func (tc *ToolContext) GetString(key string) string {
	v, ok := tc.Extra[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
