package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetQueriesInput is the input for the get_queries tool.
type GetQueriesInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the dataset"`
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
		Name:        string(ToolGetQueries),
		Description: "Get SQL queries associated with a dataset",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetQueriesInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetQueries(ctx context.Context, _ *mcp.CallToolRequest, input GetQueriesInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	queries, err := t.client.GetQueries(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(queries)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
