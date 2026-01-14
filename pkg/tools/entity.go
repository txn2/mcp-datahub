package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetEntityInput is the input for the get_entity tool.
type GetEntityInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
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
		Name:        string(ToolGetEntity),
		Description: "Get detailed metadata for a DataHub entity by its URN",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetEntityInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetEntity(ctx context.Context, _ *mcp.CallToolRequest, input GetEntityInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	entity, err := t.client.GetEntity(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(entity)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
