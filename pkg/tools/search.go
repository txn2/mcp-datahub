package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
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
		Name:        string(ToolSearch),
		Description: "Search for datasets, dashboards, pipelines, and other assets in the DataHub catalog",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleSearch(ctx context.Context, _ *mcp.CallToolRequest, input SearchInput) (*mcp.CallToolResult, any, error) {
	if input.Query == "" {
		return ErrorResult("query parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

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

	result, err := datahubClient.Search(ctx, input.Query, opts...)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(result)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
