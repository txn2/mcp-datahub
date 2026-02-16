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
		Description: t.getDescription(ToolUpdateDescription, cfg),
		Annotations: t.getAnnotations(ToolUpdateDescription, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest,
		input UpdateDescriptionInput,
	) (*mcp.CallToolResult, *UpdateDescriptionOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*UpdateDescriptionOutput); ok {
			return result, typed, err
		}
		return result, nil, err
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

	output := UpdateDescriptionOutput{
		URN:    input.URN,
		Aspect: "editableDatasetProperties",
		Action: "updated",
	}

	jsonResult, err := JSONResult(output)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, &output, nil
}
