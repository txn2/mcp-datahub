package tools

// defaultDescriptions maps each tool to its default description.
// These are used when no override is provided.
var defaultDescriptions = map[ToolName]string{
	ToolSearch: "Search for datasets, dashboards, pipelines, and other assets in the DataHub catalog. " +
		"This should be your FIRST tool when answering data questions — use it to discover " +
		"relevant datasets before querying. Results include query_context showing which datasets " +
		"are queryable in Trino and their resolved table paths. Search by topic keywords, " +
		"table names, tags, or domain concepts. Supports advanced filtering via 'filters' parameter " +
		"to search by column names (fieldPaths), column tags (fieldTags), column glossary terms " +
		"(fieldGlossaryTerms), platform, domain, owner, and more. Use 'types' to search across " +
		"multiple entity types. Follow up with datahub_get_schema or trino_describe_table for column details.",

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
		"Set level=column for column-level lineage showing which upstream columns feed each downstream column. " +
		"Default (dataset) returns dataset-level relationships with direction and depth control. " +
		"When a QueryProvider is configured, includes execution_context mapping URNs to query engine tables.",

	ToolGetQueries: "Get saved SQL queries linked to a dataset — including view definitions, common query " +
		"patterns, and example queries. For database views (v_* prefix), this returns the " +
		"actual view SQL showing all joins and transformations. Essential for understanding " +
		"how derived data is built. Also useful for showing users example query patterns.",

	ToolGetGlossaryTerm: "Get the full definition of a business glossary term and all datasets/columns linked " +
		"to it. Use when enrichment surfaces a glossary_term URN and you need the detailed " +
		"definition, or when a user asks \"what does [business term] mean?\" Returns the " +
		"canonical business definition plus all tables and columns that use this term.",

	ToolBrowse: "Browse the DataHub catalog by category. Set what=tags to list tags, " +
		"what=domains to list data domains, or what=data_products to list data products. " +
		"Use the optional filter parameter (tags only) to narrow results.",

	ToolGetDataProduct: "Get full details of a data product including its constituent datasets, owners, and " +
		"domain. Data products group related datasets for a specific business use case. " +
		"Use after datahub_browse with what=data_products to drill into a specific product " +
		"and discover all its member datasets. Useful for answering \"what data do we have about [topic]?\"",

	ToolListConnections: "List all configured DataHub server connections. " +
		"Use this to discover available connections before querying specific servers. " +
		"Pass the connection name to other tools via the 'connection' parameter.",

	// Write tools (CRUD pattern)
	ToolCreate: "Create a new entity or resource in DataHub. Set 'what' to choose the entity type: " +
		"tag, domain, glossary_term, data_product, document, application, query, incident, " +
		"structured_property, or data_contract. Returns the URN of the created entity.",

	ToolUpdate: "Update metadata on an existing DataHub entity. Set 'what' to choose what to update. " +
		"For tag/glossary_term/link/owner: 'action' is required (add or remove). " +
		"For domain/structured_properties/custom_properties: 'action' defaults to set (set or remove). " +
		"Other what values do not use 'action'. 'description' also edits tag and glossaryTerm descriptions. " +
		"Supports: description, column_description, " +
		"tag, glossary_term, link, owner, domain, structured_properties, structured_property, " +
		"custom_properties, incident_status, incident, query, document_contents, document_status, " +
		"document_related_entities, document_sub_type, and data_contract.",

	ToolDelete: "Delete an entity or resource from DataHub. Set 'what' to choose the entity type: " +
		"query, tag, domain, glossary_entity, data_product, application, document, or structured_property. " +
		"This operation is destructive and cannot be undone.",
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
