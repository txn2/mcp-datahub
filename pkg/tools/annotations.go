package tools

import "github.com/modelcontextprotocol/go-sdk/mcp"

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

// defaultAnnotations holds the default annotations for each built-in tool.
// These follow the MCP specification:
//   - ReadOnlyHint (bool, default false): tool does not modify state
//   - DestructiveHint (*bool, default true): tool may destructively update
//   - IdempotentHint (bool, default false): repeated calls produce same result
//   - OpenWorldHint (*bool, default true): tool interacts with external entities
var defaultAnnotations = map[ToolName]*mcp.ToolAnnotations{
	// Read-only tools
	ToolSearch:           {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetEntity:        {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetSchema:        {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetLineage:       {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetColumnLineage: {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetQueries:       {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetGlossaryTerm:  {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolListTags:         {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolListDomains:      {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolListDataProducts: {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolGetDataProduct:   {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolListConnections:  {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(true)},

	// Write tools
	ToolUpdateDescription:  {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolAddTag:             {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolRemoveTag:          {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolAddGlossaryTerm:    {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolRemoveGlossaryTerm: {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolAddLink:            {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
	ToolRemoveLink:         {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(true)},
}

// DefaultAnnotations returns the default annotations for a tool.
// Returns nil for unknown tool names.
func DefaultAnnotations(name ToolName) *mcp.ToolAnnotations {
	return defaultAnnotations[name]
}

// getAnnotations resolves the annotations for a tool using the priority chain:
// 1. Per-registration override (cfg.annotations) — highest priority
// 2. Toolkit-level override (t.annotations) — medium priority
// 3. Default annotations — lowest priority.
func (t *Toolkit) getAnnotations(name ToolName, cfg *toolConfig) *mcp.ToolAnnotations {
	// Per-registration override (highest priority)
	if cfg != nil && cfg.annotations != nil {
		return cfg.annotations
	}

	// Toolkit-level override (medium priority)
	if ann, ok := t.annotations[name]; ok {
		return ann
	}

	// Default annotations (lowest priority)
	return defaultAnnotations[name]
}
