package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetSchemaInput is the input for the get_schema tool.
type GetSchemaInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the dataset"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
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
		Description: t.getDescription(ToolGetSchema, cfg),
		Annotations: t.getAnnotations(ToolGetSchema, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetSchemaInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetSchema(ctx context.Context, _ *mcp.CallToolRequest, input GetSchemaInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	schema, err := datahubClient.GetSchema(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	// Build response - include table resolution if provider configured
	if t.queryProvider != nil {
		response := map[string]any{
			"schema": schema,
		}

		// Add table resolution
		if table, tableErr := t.queryProvider.ResolveTable(ctx, input.URN); tableErr == nil && table != nil {
			response["query_table"] = table
		}

		jsonResult, jsonErr := JSONResult(response)
		if jsonErr != nil {
			return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
		}
		return jsonResult, nil, nil
	}

	// No query provider - return schema only
	jsonResult, err := JSONResult(schema)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
