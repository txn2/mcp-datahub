package tools

import (
	"context"
	"encoding/json"

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
		Name:         string(ToolGetEntity),
		Description:  t.getDescription(ToolGetEntity, cfg),
		Annotations:  t.getAnnotations(ToolGetEntity, cfg),
		Icons:        t.getIcons(ToolGetEntity, cfg),
		Title:        t.getTitle(ToolGetEntity, cfg),
		OutputSchema: t.getOutputSchema(ToolGetEntity, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetEntityInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetEntity(ctx context.Context, _ *mcp.CallToolRequest, input GetEntityInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	entity, err := datahubClient.GetEntity(ctx, input.URN)
	if err != nil {
		return ErrorResult("GetEntity failed for " + input.URN + ": " + err.Error()), nil, nil
	}

	if entity == nil {
		return ErrorResult("GetEntity returned nil for " + input.URN), nil, nil
	}

	if t.queryProvider != nil {
		return t.enrichEntityWithQueryContext(ctx, entity, input.URN)
	}

	return formatJSONResult(entity)
}

// enrichEntityWithQueryContext flattens entity fields to top level and appends
// query provider data at the same level (matches OutputSchema).
func (t *Toolkit) enrichEntityWithQueryContext(ctx context.Context, entity any, urn string) (*mcp.CallToolResult, any, error) {
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		return ErrorResult("failed to marshal entity: " + err.Error()), nil, nil
	}
	response := map[string]any{}
	if err = json.Unmarshal(entityJSON, &response); err != nil {
		return ErrorResult("failed to build response: " + err.Error()), nil, nil
	}

	if table, tableErr := t.queryProvider.ResolveTable(ctx, urn); tableErr == nil && table != nil {
		response["query_table"] = table.String()
	}
	if examples, examplesErr := t.queryProvider.GetQueryExamples(ctx, urn); examplesErr == nil && len(examples) > 0 {
		response["query_examples"] = examples
	}
	if avail, availErr := t.queryProvider.GetTableAvailability(ctx, urn); availErr == nil && avail != nil {
		response["query_availability"] = avail
	}

	return formatJSONResult(response)
}
