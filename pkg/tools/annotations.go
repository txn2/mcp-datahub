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
	ToolSearch:           {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetEntity:        {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetSchema:        {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetLineage:       {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetColumnLineage: {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetQueries:       {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetGlossaryTerm:  {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolListTags:         {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolListDomains:      {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolListDataProducts: {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolGetDataProduct:   {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolListConnections:  {ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: boolPtr(false)},

	// Write tools
	ToolUpdateDescription:  {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolAddTag:             {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolRemoveTag:          {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolAddGlossaryTerm:    {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolRemoveGlossaryTerm: {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolAddLink:            {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
	ToolRemoveLink:         {DestructiveHint: boolPtr(false), IdempotentHint: true, OpenWorldHint: boolPtr(false)},
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
