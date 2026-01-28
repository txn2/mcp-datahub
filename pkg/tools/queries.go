package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetQueriesInput is the input for the get_queries tool.
type GetQueriesInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the dataset"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetQueriesTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		queriesInput, ok := input.(GetQueriesInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetQueries(ctx, req, queriesInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetQueries, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name: string(ToolGetQueries),
		Description: "Get saved SQL queries associated with a dataset. Returns Query entities " +
			"linked to this dataset, including query text, description, and creator information.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetQueriesInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetQueries(ctx context.Context, _ *mcp.CallToolRequest, input GetQueriesInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	queries, err := datahubClient.GetQueries(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(queries)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
