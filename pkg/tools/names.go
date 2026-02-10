package tools

// ToolName identifies a DataHub MCP tool.
type ToolName string

// Tool name constants.
const (
	ToolSearch           ToolName = "datahub_search"
	ToolGetEntity        ToolName = "datahub_get_entity"
	ToolGetSchema        ToolName = "datahub_get_schema"
	ToolGetLineage       ToolName = "datahub_get_lineage"
	ToolGetColumnLineage ToolName = "datahub_get_column_lineage"
	ToolGetQueries       ToolName = "datahub_get_queries"
	ToolGetGlossaryTerm  ToolName = "datahub_get_glossary_term"
	ToolListTags         ToolName = "datahub_list_tags"
	ToolListDomains      ToolName = "datahub_list_domains"
	ToolListDataProducts ToolName = "datahub_list_data_products"
	ToolGetDataProduct   ToolName = "datahub_get_data_product"
	ToolListConnections  ToolName = "datahub_list_connections"

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
func AllTools() []ToolName {
	return []ToolName{
		ToolSearch,
		ToolGetEntity,
		ToolGetSchema,
		ToolGetLineage,
		ToolGetColumnLineage,
		ToolGetQueries,
		ToolGetGlossaryTerm,
		ToolListTags,
		ToolListDomains,
		ToolListDataProducts,
		ToolGetDataProduct,
		ToolListConnections,
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
