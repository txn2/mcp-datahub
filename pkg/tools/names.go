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

	// Deprecated: use ToolBrowse with what="tags". Kept as alias for one release cycle.
	ToolListTags ToolName = "datahub_list_tags"
	// Deprecated: use ToolBrowse with what="domains". Kept as alias for one release cycle.
	ToolListDomains ToolName = "datahub_list_domains"
	// Deprecated: use ToolBrowse with what="data_products". Kept as alias for one release cycle.
	ToolListDataProducts ToolName = "datahub_list_data_products"
	// Deprecated: use ToolGetLineage with level="column". Kept as alias for one release cycle.
	ToolGetColumnLineage ToolName = "datahub_get_column_lineage"

	// Write tool names.
	ToolUpdateDescription  ToolName = "datahub_update_description"
	ToolAddTag             ToolName = "datahub_add_tag"
	ToolRemoveTag          ToolName = "datahub_remove_tag"
	ToolAddGlossaryTerm    ToolName = "datahub_add_glossary_term"
	ToolRemoveGlossaryTerm ToolName = "datahub_remove_glossary_term"
	ToolAddLink            ToolName = "datahub_add_link"
	ToolRemoveLink         ToolName = "datahub_remove_link"
)

// AllTools returns all available read-only tool names.
// This does not include write tools for backward compatibility.
// Deprecated aliases are not included; they are registered
// separately via DeprecatedTools().
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

// DeprecatedTools returns tool names kept as aliases for one release cycle.
func DeprecatedTools() []ToolName {
	return []ToolName{
		ToolListTags,
		ToolListDomains,
		ToolListDataProducts,
		ToolGetColumnLineage,
	}
}

// WriteTools returns all write tool names.
func WriteTools() []ToolName {
	return []ToolName{
		ToolUpdateDescription,
		ToolAddTag,
		ToolRemoveTag,
		ToolAddGlossaryTerm,
		ToolRemoveGlossaryTerm,
		ToolAddLink,
		ToolRemoveLink,
	}
}
