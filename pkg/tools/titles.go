package tools

// defaultTitles maps each tool to its default human-readable display name.
// These are used by MCP clients (e.g., Claude Desktop) to display tools in the UI.
// Display name precedence in the SDK: Tool.Title > ToolAnnotations.Title > Tool.Name.
var defaultTitles = map[ToolName]string{
	// Read tools
	ToolSearch:          "Search Catalog",
	ToolGetEntity:       "Get Entity",
	ToolGetSchema:       "Get Schema",
	ToolGetLineage:      "Get Lineage",
	ToolGetQueries:      "Get Queries",
	ToolBrowse:          "Browse Catalog",
	ToolGetGlossaryTerm: "Get Glossary Term",
	ToolGetDataProduct:  "Get Data Product",
	ToolListConnections: "List Connections",

	// Write tools
	ToolUpdateDescription:  "Update Description",
	ToolAddTag:             "Add Tag",
	ToolRemoveTag:          "Remove Tag",
	ToolAddGlossaryTerm:    "Add Glossary Term",
	ToolRemoveGlossaryTerm: "Remove Glossary Term",
	ToolAddLink:            "Add Link",
	ToolRemoveLink:         "Remove Link",
	ToolCreateDocument:     "Create Document",
	ToolUpdateDocument:     "Update Document",
}

// DefaultTitle returns the default human-readable title for a tool.
// Returns an empty string if the tool name is not recognized.
func DefaultTitle(name ToolName) string {
	return defaultTitles[name]
}

// getTitle resolves the title for a tool using the priority chain:
//  1. Per-registration cfg.title (highest)
//  2. Toolkit-level t.titles map
//  3. defaultTitles map (lowest/default)
func (t *Toolkit) getTitle(name ToolName, cfg *toolConfig) string {
	// Highest priority: per-registration override
	if cfg != nil && cfg.title != nil {
		return *cfg.title
	}

	// Middle priority: toolkit-level override
	if title, ok := t.titles[name]; ok {
		return title
	}

	// Lowest priority: default
	return defaultTitles[name]
}
