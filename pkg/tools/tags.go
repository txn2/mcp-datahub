package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListTagsInput is the input for the list_tags tool.
type ListTagsInput struct {
	Filter string `json:"filter,omitempty" jsonschema_description:"Optional filter string to match tag names"`
}

func (t *Toolkit) registerListTagsTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		tagsInput, ok := input.(ListTagsInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleListTags(ctx, req, tagsInput)
	}

	wrappedHandler := t.wrapHandler(ToolListTags, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolListTags),
		Description: "List available tags in the DataHub catalog",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListTagsInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleListTags(ctx context.Context, _ *mcp.CallToolRequest, input ListTagsInput) (*mcp.CallToolResult, any, error) {
	tags, err := t.client.ListTags(ctx, input.Filter)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(tags)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
