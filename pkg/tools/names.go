package tools

// ToolName identifies a DataHub MCP tool.
type ToolName string

// Tool name constants.
const (
	ToolSearch          ToolName = "datahub_search"
	ToolGetEntity       ToolName = "datahub_get_entity"
	ToolGetSchema       ToolName = "datahub_get_schema"
	ToolGetLineage      ToolName = "datahub_get_lineage"
	ToolGetQueries      ToolName = "datahub_get_queries"
	ToolBrowse          ToolName = "datahub_browse"
	ToolGetGlossaryTerm ToolName = "datahub_get_glossary_term"
	ToolGetDataProduct  ToolName = "datahub_get_data_product"
	ToolListConnections ToolName = "datahub_list_connections"

	// Write tool names (CRUD pattern).
	ToolCreate ToolName = "datahub_create"
	ToolUpdate ToolName = "datahub_update"
	ToolDelete ToolName = "datahub_delete"
)

// AllTools returns all available read-only tool names.
// This does not include write tools for backward compatibility.
func AllTools() []ToolName {
	return []ToolName{
		ToolSearch,
		ToolGetEntity,
		ToolGetSchema,
		ToolGetLineage,
		ToolGetQueries,
		ToolBrowse,
		ToolGetGlossaryTerm,
		ToolGetDataProduct,
		ToolListConnections,
	}
}

// WriteTools returns all write tool names.
func WriteTools() []ToolName {
	return []ToolName{
		ToolCreate,
		ToolUpdate,
		ToolDelete,
	}
}
