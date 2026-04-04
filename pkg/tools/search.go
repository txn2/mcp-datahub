package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// SearchInput is the input for the search tool.
type SearchInput struct {
	Query string `json:"query" jsonschema_description:"Search query string"`
	// EntityType: DATASET, DASHBOARD, DATA_FLOW, DATA_JOB, CONTAINER, etc.
	EntityType string `json:"entity_type,omitempty" jsonschema_description:"Entity type (single). Defaults to DATASET."`
	// Types allows searching across multiple entity types. Overrides entity_type.
	Types []string `json:"types,omitempty" jsonschema_description:"Entity types to search across. Overrides entity_type."`
	// Filters enable advanced field-level filtering via searchAcrossEntities.
	Filters []SearchFilterInput `json:"filters,omitempty" jsonschema_description:"Advanced field-level filters (AND'd together)."`
	Limit   int                 `json:"limit,omitempty" jsonschema_description:"Maximum number of results (default: 10, max: 100)"`
	Offset  int                 `json:"offset,omitempty" jsonschema_description:"Result offset for pagination"`
	// Mode selects the search strategy: "keyword" (default) or "semantic".
	// Semantic search uses vector embeddings for natural language queries.
	// Requires DataHub 1.4.x with OpenSearch 2.19.3+.
	Mode string `json:"mode,omitempty" jsonschema_description:"Search mode: keyword (default) or semantic"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// SearchFilterInput represents a single filter criterion for advanced search.
type SearchFilterInput struct {
	// Field is the filter field name.
	Field string `json:"field" jsonschema_description:"Filter field name (see tool description for options)"`
	// Value is a convenience field for a single value. Use values for multiple.
	Value string `json:"value,omitempty" jsonschema_description:"Single filter value. Use 'values' for multiple."`
	// Values are the values to match against.
	Values []string `json:"values,omitempty" jsonschema_description:"Filter values to match against"`
	// Condition is the filter operator. Defaults to EQUAL if omitted.
	Condition string `json:"condition,omitempty" jsonschema_description:"Filter operator: CONTAIN, EQUAL (default), IN, EXISTS"`
	// Negated inverts the filter to exclude matches.
	Negated bool `json:"negated,omitempty" jsonschema_description:"If true, exclude entities matching this filter"`
}

func (t *Toolkit) registerSearchTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		searchInput, ok := input.(SearchInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleSearch(ctx, req, searchInput)
	}

	wrappedHandler := t.wrapHandler(ToolSearch, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolSearch),
		Description:  t.getDescription(ToolSearch, cfg),
		Annotations:  t.getAnnotations(ToolSearch, cfg),
		Icons:        t.getIcons(ToolSearch, cfg),
		Title:        t.getTitle(ToolSearch, cfg),
		OutputSchema: t.getOutputSchema(ToolSearch, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

// buildSearchOptions constructs base SearchOptions (limit/offset) from input parameters.
func buildSearchOptions(input SearchInput) []client.SearchOption {
	var opts []client.SearchOption
	if input.Limit > 0 {
		opts = append(opts, client.WithLimit(input.Limit))
	}
	if input.Offset > 0 {
		opts = append(opts, client.WithOffset(input.Offset))
	}
	return opts
}

// resolveSearchTypes determines the entity types to search.
// Priority: types > entity_type > default (DATASET).
func resolveSearchTypes(input SearchInput) []string {
	if len(input.Types) > 0 {
		return input.Types
	}
	if input.EntityType != "" {
		return []string{input.EntityType}
	}
	return []string{"DATASET"}
}

// validateFilters checks that all filter inputs have a non-empty field and at least one value.
func validateFilters(filters []SearchFilterInput) error {
	for i, f := range filters {
		if f.Field == "" {
			return fmt.Errorf("filter[%d]: field is required", i)
		}
		if f.Value == "" && len(f.Values) == 0 {
			return fmt.Errorf("filter[%d] (%s): value or values is required", i, f.Field)
		}
	}
	return nil
}

// convertFilters converts tool-layer filter inputs to client-layer SearchFilter values.
func convertFilters(inputs []SearchFilterInput) []client.SearchFilter {
	filters := make([]client.SearchFilter, 0, len(inputs))
	for _, f := range inputs {
		values := f.Values
		if f.Value != "" {
			values = append([]string{f.Value}, values...)
		}
		filters = append(filters, client.SearchFilter{
			Field:     f.Field,
			Values:    values,
			Condition: f.Condition,
			Negated:   f.Negated,
		})
	}
	return filters
}

func (t *Toolkit) handleSearch(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
	if input.Query == "" {
		return ErrorResult("query parameter is required"), nil, nil
	}

	if input.Mode != "" && input.Mode != "keyword" && input.Mode != "semantic" {
		return ErrorResult("invalid mode: must be 'keyword' or 'semantic'"), nil, nil
	}

	if err := validateFilters(input.Filters); err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	opts := buildSearchOptions(input)
	searchTypes := resolveSearchTypes(input)
	opts = append(opts, client.WithTypes(searchTypes))

	if len(input.Filters) > 0 {
		opts = append(opts, client.WithSearchFilters(convertFilters(input.Filters)))
	}

	var result *types.SearchResult
	if input.Mode == "semantic" {
		result, err = datahubClient.SemanticSearch(ctx, input.Query, opts...)
	} else {
		result, err = datahubClient.SearchAcrossEntities(ctx, input.Query, opts...)
	}
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	return t.formatSearchResult(ctx, result)
}

// formatSearchResult formats search results, enriching with query context if available.
// SearchResult contains only concrete types (no any fields), so direct field mapping
// is used instead of a json roundtrip — no marshal error path is possible.
func (t *Toolkit) formatSearchResult(ctx context.Context, result *types.SearchResult) (*mcp.CallToolResult, any, error) {
	queryContext := t.buildQueryContext(ctx, result)

	if len(queryContext) > 0 {
		// Build flat response matching OutputSchema (entities/total at top level)
		response := map[string]any{
			"total":    result.Total,
			"entities": result.Entities,
			"offset":   result.Offset,
			"limit":    result.Limit,
		}
		response["query_context"] = queryContext
		return formatJSONResult(response)
	}

	return formatJSONResult(result)
}

// buildQueryContext builds query availability context for search results.
func (t *Toolkit) buildQueryContext(ctx context.Context, result *types.SearchResult) map[string]any {
	if t.queryProvider == nil || len(result.Entities) == 0 {
		return nil
	}

	queryContext := make(map[string]any)
	for _, entity := range result.Entities {
		avail, err := t.queryProvider.GetTableAvailability(ctx, entity.URN)
		if err != nil || avail == nil {
			continue
		}
		entityCtx := map[string]any{"available": avail.Available}
		if avail.Table != nil {
			entityCtx["table"] = avail.Table.String()
		}
		queryContext[entity.URN] = entityCtx
	}
	return queryContext
}

// formatJSONResult is a helper to format and return JSON results.
// data is returned as-is as the structured output second return value so that
// go-sdk populates structuredContent when the tool declares an outputSchema.
func formatJSONResult(data any) (*mcp.CallToolResult, any, error) {
	jsonResult, err := JSONResult(data)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, data, nil
}
