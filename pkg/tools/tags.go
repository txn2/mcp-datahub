package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListTagsInput is the input for the list_tags tool.
type ListTagsInput struct {
	Filter string `json:"filter,omitempty" jsonschema_description:"Optional filter string to match tag names"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
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
		Name:         string(ToolListTags),
		Description:  t.getDescription(ToolListTags, cfg),
		Annotations:  t.getAnnotations(ToolListTags, cfg),
		Icons:        t.getIcons(ToolListTags, cfg),
		Title:        t.getTitle(ToolListTags, cfg),
		OutputSchema: t.getOutputSchema(ToolListTags, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListTagsInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleListTags(ctx context.Context, _ *mcp.CallToolRequest, input ListTagsInput) (*mcp.CallToolResult, any, error) {
	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	tags, err := datahubClient.ListTags(ctx, input.Filter)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	output := ListTagsOutput{Tags: tags}
	jsonResult, err := JSONResult(output)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, &output, nil
}
