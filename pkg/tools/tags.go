package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListTagsInput is the input for the deprecated datahub_list_tags tool.
//
// Deprecated: use BrowseInput with What="tags" instead.
type ListTagsInput struct {
	Filter string `json:"filter,omitempty" jsonschema_description:"Optional filter string to match tag names"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// registerListTagsTool registers the deprecated datahub_list_tags alias.
// It delegates to handleBrowse with What="tags".
func (t *Toolkit) registerListTagsTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		tagsInput, ok := input.(ListTagsInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		browseInput := BrowseInput{
			What:       "tags",
			Filter:     tagsInput.Filter,
			Connection: tagsInput.Connection,
		}
		return t.handleBrowse(ctx, req, browseInput)
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
