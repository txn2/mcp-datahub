package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetSchemaInput is the input for the get_schema tool.
type GetSchemaInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the dataset"`
}

func (t *Toolkit) registerGetSchemaTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		schemaInput, ok := input.(GetSchemaInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetSchema(ctx, req, schemaInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetSchema, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolGetSchema),
		Description: "Get the schema (fields, types, descriptions) for a dataset",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSchemaInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetSchema(ctx context.Context, _ *mcp.CallToolRequest, input GetSchemaInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	schema, err := t.client.GetSchema(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(schema)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
