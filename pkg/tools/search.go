package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// SearchInput is the input for the search tool.
type SearchInput struct {
	Query string `json:"query" jsonschema_description:"Search query string"`
	// EntityType: DATASET, DASHBOARD, DATA_FLOW, DATA_JOB, CONTAINER, TAG, GLOSSARY_TERM, DATA_PRODUCT, etc.
	EntityType string `json:"entity_type,omitempty" jsonschema_description:"Entity type to search. Defaults to DATASET."`
	Limit      int    `json:"limit,omitempty" jsonschema_description:"Maximum number of results (default: 10, max: 100)"`
	Offset     int    `json:"offset,omitempty" jsonschema_description:"Result offset for pagination"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
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

// buildSearchOptions constructs SearchOptions from input parameters.
func buildSearchOptions(input SearchInput) []client.SearchOption {
	var opts []client.SearchOption
	if input.EntityType != "" {
		opts = append(opts, client.WithEntityType(input.EntityType))
	}
	if input.Limit > 0 {
		opts = append(opts, client.WithLimit(input.Limit))
	}
	if input.Offset > 0 {
		opts = append(opts, client.WithOffset(input.Offset))
	}
	return opts
}

func (t *Toolkit) handleSearch(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
	if input.Query == "" {
		return ErrorResult("query parameter is required"), nil, nil
	}

	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	result, err := datahubClient.Search(ctx, input.Query, buildSearchOptions(input)...)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	return t.formatSearchResult(ctx, result)
}

// formatSearchResult formats search results, enriching with query context if available.
func (t *Toolkit) formatSearchResult(ctx context.Context, result *types.SearchResult) (*mcp.CallToolResult, any, error) {
	queryContext := t.buildQueryContext(ctx, result)

	if len(queryContext) > 0 {
		response := map[string]any{
			"result":        result,
			"query_context": queryContext,
		}
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
