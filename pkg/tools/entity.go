package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetEntityInput is the input for the get_entity tool.
type GetEntityInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetEntityTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		entityInput, ok := input.(GetEntityInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetEntity(ctx, req, entityInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetEntity, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name: string(ToolGetEntity),
		Description: "Get detailed metadata for a DataHub entity by its URN. " +
			"When a QueryProvider (e.g., Trino) is configured, also returns: " +
			"query_table (resolved table path), query_examples (auto-generated SQL), " +
			"query_availability (row count, availability status).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetEntityInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetEntity(ctx context.Context, _ *mcp.CallToolRequest, input GetEntityInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	entity, err := datahubClient.GetEntity(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	// Build response - include query context if provider configured
	if t.queryProvider != nil {
		response := map[string]any{
			"entity": entity,
		}

		// Add table resolution
		if table, tableErr := t.queryProvider.ResolveTable(ctx, input.URN); tableErr == nil && table != nil {
			response["query_table"] = table
		}

		// Add query examples
		if examples, examplesErr := t.queryProvider.GetQueryExamples(ctx, input.URN); examplesErr == nil && len(examples) > 0 {
			response["query_examples"] = examples
		}

		// Add availability status
		if avail, availErr := t.queryProvider.GetTableAvailability(ctx, input.URN); availErr == nil && avail != nil {
			response["query_availability"] = avail
		}

		jsonResult, jsonErr := JSONResult(response)
		if jsonErr != nil {
			return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
		}
		return jsonResult, nil, nil
	}

	// No query provider - return entity only
	jsonResult, err := JSONResult(entity)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
