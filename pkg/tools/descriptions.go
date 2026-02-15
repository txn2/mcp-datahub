package tools

// defaultDescriptions maps each tool to its default description.
// These are used when no override is provided.
var defaultDescriptions = map[ToolName]string{
	ToolSearch: "Search for datasets, dashboards, pipelines, and other assets in the DataHub catalog. " +
		"This should be your FIRST tool when answering data questions — use it to discover " +
		"relevant datasets before querying. Results include query_context showing which datasets " +
		"are queryable in Trino and their resolved table paths. Search by topic keywords, " +
		"table names, tags, or domain concepts. Follow up with datahub_get_schema or " +
		"trino_describe_table for column details.",

	ToolGetEntity: "Get comprehensive metadata for a DataHub entity including description, owners, tags, " +
		"glossary terms, domain, deprecation status, quality score, and custom properties. " +
		"Use this when you need the FULL metadata picture for a specific dataset — especially " +
		"ownership, quality scores, and deprecation warnings. Returns more metadata fields than " +
		"get_schema (which focuses on columns). Also includes query_table path and row count " +
		"when a query engine is configured.",

	ToolGetSchema: "Get the schema (fields, types, descriptions) for a dataset. " +
		"Returns query_table (resolved table path) when QueryProvider is configured. " +
		"For row counts and query examples, use datahub_get_entity instead.",

	ToolGetLineage: "Get upstream or downstream lineage for a DataHub entity. " +
		"When a QueryProvider is configured, includes execution_context " +
		"mapping URNs to query engine tables.",

	ToolGetColumnLineage: "Get column-level lineage showing exactly which upstream columns feed each downstream " +
		"column. Use this when a user asks \"where does this column come from?\" or when you " +
		"need to trace a specific metric through transformations. More precise than " +
		"datahub_get_lineage which shows dataset-level relationships. Essential for debugging " +
		"data quality issues in derived tables and views.",

	ToolGetQueries: "Get saved SQL queries linked to a dataset — including view definitions, common query " +
		"patterns, and example queries. For database views (v_* prefix), this returns the " +
		"actual view SQL showing all joins and transformations. Essential for understanding " +
		"how derived data is built. Also useful for showing users example query patterns.",

	ToolGetGlossaryTerm: "Get the full definition of a business glossary term and all datasets/columns linked " +
		"to it. Use when enrichment surfaces a glossary_term URN and you need the detailed " +
		"definition, or when a user asks \"what does [business term] mean?\" Returns the " +
		"canonical business definition plus all tables and columns that use this term.",

	ToolListTags:         "List available tags in the DataHub catalog",
	ToolListDomains:      "List data domains in the DataHub catalog",
	ToolListDataProducts: "List data products in the DataHub catalog. Data products group datasets for specific business use cases.",

	ToolGetDataProduct: "Get full details of a data product including its constituent datasets, owners, and " +
		"domain. Data products group related datasets for a specific business use case. " +
		"Use after datahub_list_data_products to drill into a specific product and discover " +
		"all its member datasets. Useful for answering \"what data do we have about [topic]?\"",

	ToolListConnections: "List all configured DataHub server connections. " +
		"Use this to discover available connections before querying specific servers. " +
		"Pass the connection name to other tools via the 'connection' parameter.",

	// Write tools
	ToolUpdateDescription:  "Update the description of a DataHub entity",
	ToolAddTag:             "Add a tag to a DataHub entity",
	ToolRemoveTag:          "Remove a tag from a DataHub entity",
	ToolAddGlossaryTerm:    "Add a glossary term to a DataHub entity",
	ToolRemoveGlossaryTerm: "Remove a glossary term from a DataHub entity",
	ToolAddLink:            "Add a link to a DataHub entity",
	ToolRemoveLink:         "Remove a link from a DataHub entity",
}

// DefaultDescription returns the default description for a tool.
// Returns an empty string if the tool name is not recognized.
func DefaultDescription(name ToolName) string {
	return defaultDescriptions[name]
}

// getDescription returns the description for a tool with three-tier priority:
//  1. Per-registration cfg.description (highest)
//  2. Toolkit-level t.descriptions map
//  3. defaultDescriptions map (lowest/default)
func (t *Toolkit) getDescription(name ToolName, cfg *toolConfig) string {
	// Highest priority: per-registration override
	if cfg != nil && cfg.description != nil {
		return *cfg.description
	}

	// Middle priority: toolkit-level override
	if desc, ok := t.descriptions[name]; ok {
		return desc
	}

	// Lowest priority: default
	return defaultDescriptions[name]
}
