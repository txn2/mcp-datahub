package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddTagInput is the input for the add_tag tool.
type AddTagInput struct {
	URN        string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	TagURN     string `json:"tag_urn" jsonschema_description:"The URN of the tag to add (e.g., urn:li:tag:PII)"`
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// RemoveTagInput is the input for the remove_tag tool.
type RemoveTagInput struct {
	URN        string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	TagURN     string `json:"tag_urn" jsonschema_description:"The URN of the tag to remove (e.g., urn:li:tag:PII)"`
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerAddTagTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		tagInput, ok := input.(AddTagInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleAddTag(ctx, req, tagInput)
	}

	wrappedHandler := t.wrapHandler(ToolAddTag, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolAddTag),
		Description: t.getDescription(ToolAddTag, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddTagInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) registerRemoveTagTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		tagInput, ok := input.(RemoveTagInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleRemoveTag(ctx, req, tagInput)
	}

	wrappedHandler := t.wrapHandler(ToolRemoveTag, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolRemoveTag),
		Description: t.getDescription(ToolRemoveTag, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RemoveTagInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleAddTag(ctx context.Context, _ *mcp.CallToolRequest, input AddTagInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}
	if input.TagURN == "" {
		return ErrorResult("tag_urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.AddTag(ctx, input.URN, input.TagURN)
	if err != nil {
		return ErrorResult("AddTag failed: " + err.Error()), nil, nil
	}

	result := map[string]string{
		"urn":    input.URN,
		"tag":    input.TagURN,
		"aspect": "globalTags",
		"action": "added",
	}

	jsonResult, err := JSONResult(result)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, nil, nil
}

func (t *Toolkit) handleRemoveTag(ctx context.Context, _ *mcp.CallToolRequest, input RemoveTagInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}
	if input.TagURN == "" {
		return ErrorResult("tag_urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.RemoveTag(ctx, input.URN, input.TagURN)
	if err != nil {
		return ErrorResult("RemoveTag failed: " + err.Error()), nil, nil
	}

	result := map[string]string{
		"urn":    input.URN,
		"tag":    input.TagURN,
		"aspect": "globalTags",
		"action": "removed",
	}

	jsonResult, err := JSONResult(result)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, nil, nil
}
