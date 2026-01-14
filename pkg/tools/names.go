package tools

// ToolName identifies a DataHub MCP tool.
type ToolName string

// Tool name constants.
const (
	ToolSearch           ToolName = "datahub_search"
	ToolGetEntity        ToolName = "datahub_get_entity"
	ToolGetSchema        ToolName = "datahub_get_schema"
	ToolGetLineage       ToolName = "datahub_get_lineage"
	ToolGetQueries       ToolName = "datahub_get_queries"
	ToolGetGlossaryTerm  ToolName = "datahub_get_glossary_term"
	ToolListTags         ToolName = "datahub_list_tags"
	ToolListDomains      ToolName = "datahub_list_domains"
	ToolListDataProducts ToolName = "datahub_list_data_products"
	ToolGetDataProduct   ToolName = "datahub_get_data_product"
)

// AllTools returns all available tool names.
func AllTools() []ToolName {
	return []ToolName{
		ToolSearch,
		ToolGetEntity,
		ToolGetSchema,
		ToolGetLineage,
		ToolGetQueries,
		ToolGetGlossaryTerm,
		ToolListTags,
		ToolListDomains,
		ToolListDataProducts,
		ToolGetDataProduct,
	}
}
