package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// UpdateDescriptionInput is the input for the update_description tool.
type UpdateDescriptionInput struct {
	URN         string `json:"urn" jsonschema_description:"The DataHub URN of the entity to update"`
	Description string `json:"description" jsonschema_description:"The new description text"`
	Connection  string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerUpdateDescriptionTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		descInput, ok := input.(UpdateDescriptionInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleUpdateDescription(ctx, req, descInput)
	}

	wrappedHandler := t.wrapHandler(ToolUpdateDescription, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolUpdateDescription),
		Description: "Update the description of a DataHub entity",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateDescriptionInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleUpdateDescription(
	ctx context.Context, _ *mcp.CallToolRequest, input UpdateDescriptionInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.UpdateDescription(ctx, input.URN, input.Description)
	if err != nil {
		return ErrorResult("UpdateDescription failed: " + err.Error()), nil, nil
	}

	result := map[string]string{
		"urn":    input.URN,
		"aspect": "editableDatasetProperties",
		"action": "updated",
	}

	jsonResult, err := JSONResult(result)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, nil, nil
}
