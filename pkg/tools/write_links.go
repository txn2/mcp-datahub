package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddLinkInput is the input for the add_link tool.
type AddLinkInput struct {
	URN         string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	URL         string `json:"url" jsonschema_description:"The URL of the link to add"`
	Description string `json:"description" jsonschema_description:"A description of the link"`
	Connection  string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// RemoveLinkInput is the input for the remove_link tool.
type RemoveLinkInput struct {
	URN        string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	URL        string `json:"url" jsonschema_description:"The URL of the link to remove"`
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerAddLinkTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		linkInput, ok := input.(AddLinkInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleAddLink(ctx, req, linkInput)
	}

	wrappedHandler := t.wrapHandler(ToolAddLink, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolAddLink),
		Description:  t.getDescription(ToolAddLink, cfg),
		Annotations:  t.getAnnotations(ToolAddLink, cfg),
		Icons:        t.getIcons(ToolAddLink, cfg),
		Title:        t.getTitle(ToolAddLink, cfg),
		OutputSchema: t.getOutputSchema(ToolAddLink, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddLinkInput) (*mcp.CallToolResult, *AddLinkOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*AddLinkOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) registerRemoveLinkTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		linkInput, ok := input.(RemoveLinkInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleRemoveLink(ctx, req, linkInput)
	}

	wrappedHandler := t.wrapHandler(ToolRemoveLink, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolRemoveLink),
		Description:  t.getDescription(ToolRemoveLink, cfg),
		Annotations:  t.getAnnotations(ToolRemoveLink, cfg),
		Icons:        t.getIcons(ToolRemoveLink, cfg),
		Title:        t.getTitle(ToolRemoveLink, cfg),
		OutputSchema: t.getOutputSchema(ToolRemoveLink, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RemoveLinkInput) (*mcp.CallToolResult, *RemoveLinkOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*RemoveLinkOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleAddLink(ctx context.Context, _ *mcp.CallToolRequest, input AddLinkInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}
	if input.URL == "" {
		return ErrorResult("url parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.AddLink(ctx, input.URN, input.URL, input.Description)
	if err != nil {
		return ErrorResult("AddLink failed: " + err.Error()), nil, nil
	}

	output := AddLinkOutput{
		URN:    input.URN,
		URL:    input.URL,
		Aspect: "institutionalMemory",
		Action: "added",
	}

	jsonResult, err := JSONResult(output)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, &output, nil
}

func (t *Toolkit) handleRemoveLink(ctx context.Context, _ *mcp.CallToolRequest, input RemoveLinkInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}
	if input.URL == "" {
		return ErrorResult("url parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.RemoveLink(ctx, input.URN, input.URL)
	if err != nil {
		return ErrorResult("RemoveLink failed: " + err.Error()), nil, nil
	}

	output := RemoveLinkOutput{
		URN:    input.URN,
		URL:    input.URL,
		Aspect: "institutionalMemory",
		Action: "removed",
	}

	jsonResult, err := JSONResult(output)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, &output, nil
}
