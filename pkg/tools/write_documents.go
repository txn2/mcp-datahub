package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// CreateDocumentInput is the input for the datahub_create_document tool.
type CreateDocumentInput struct {
	Title   string `json:"title" jsonschema_description:"Document title"`
	Content string `json:"content" jsonschema_description:"Document body text (markdown supported)"`
	//nolint:lll // struct tag cannot be split
	State         string   `json:"state,omitempty" jsonschema_description:"Initial state: PUBLISHED (default) or UNPUBLISHED" jsonschema_enum:"PUBLISHED,UNPUBLISHED"`
	RelatedAssets []string `json:"related_assets,omitempty" jsonschema_description:"Entity URNs to link to this document"`
	//nolint:lll // struct tag cannot be split
	GlobalContext *bool  `json:"global_context,omitempty" jsonschema_description:"Show in global search (default true; false for entity-specific AI context)"`
	Connection    string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// UpdateDocumentInput is the input for the datahub_update_document tool.
type UpdateDocumentInput struct {
	URN           string   `json:"urn" jsonschema_description:"The document URN (urn:li:document:{id})"`
	Title         string   `json:"title,omitempty" jsonschema_description:"New title (omit to keep current)"`
	Content       string   `json:"content,omitempty" jsonschema_description:"New body text (omit to keep current)"`
	State         string   `json:"state,omitempty" jsonschema_description:"New state: PUBLISHED or UNPUBLISHED (omit to keep current)"`
	RelatedAssets []string `json:"related_assets,omitempty" jsonschema_description:"Replace linked entity URNs (omit to keep current)"`
	Connection    string   `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerCreateDocumentTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		docInput, ok := input.(CreateDocumentInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleCreateDocument(ctx, req, docInput)
	}

	wrappedHandler := t.wrapHandler(ToolCreateDocument, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolCreateDocument),
		Description:  t.getDescription(ToolCreateDocument, cfg),
		Annotations:  t.getAnnotations(ToolCreateDocument, cfg),
		Icons:        t.getIcons(ToolCreateDocument, cfg),
		Title:        t.getTitle(ToolCreateDocument, cfg),
		OutputSchema: t.getOutputSchema(ToolCreateDocument, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest,
		input CreateDocumentInput,
	) (*mcp.CallToolResult, *CreateDocumentOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*CreateDocumentOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleCreateDocument(
	ctx context.Context, _ *mcp.CallToolRequest, input CreateDocumentInput,
) (*mcp.CallToolResult, any, error) {
	if input.Title == "" {
		return ErrorResult("title parameter is required"), nil, nil
	}
	if input.Content == "" {
		return ErrorResult("content parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	createInput := types.CreateDocumentInput{
		Title:               input.Title,
		Content:             input.Content,
		State:               input.State,
		RelatedAssets:       input.RelatedAssets,
		ShowInGlobalContext: input.GlobalContext,
	}

	urn, createErr := datahubClient.CreateDocument(ctx, createInput)
	if createErr != nil {
		return ErrorResult("CreateDocument failed: " + createErr.Error()), nil, nil
	}

	output := CreateDocumentOutput{
		URN:    urn,
		Action: "created",
	}

	jsonResult, jsonErr := JSONResult(output)
	if jsonErr != nil {
		return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
	}
	return jsonResult, &output, nil
}

func (t *Toolkit) registerUpdateDocumentTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		docInput, ok := input.(UpdateDocumentInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleUpdateDocument(ctx, req, docInput)
	}

	wrappedHandler := t.wrapHandler(ToolUpdateDocument, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolUpdateDocument),
		Description:  t.getDescription(ToolUpdateDocument, cfg),
		Annotations:  t.getAnnotations(ToolUpdateDocument, cfg),
		Icons:        t.getIcons(ToolUpdateDocument, cfg),
		Title:        t.getTitle(ToolUpdateDocument, cfg),
		OutputSchema: t.getOutputSchema(ToolUpdateDocument, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest,
		input UpdateDocumentInput,
	) (*mcp.CallToolResult, *UpdateDocumentOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*UpdateDocumentOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleUpdateDocument(
	ctx context.Context, _ *mcp.CallToolRequest, input UpdateDocumentInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	// Update contents (title and/or content)
	if input.Title != "" || input.Content != "" {
		if contentsErr := datahubClient.UpdateDocumentContents(ctx, input.URN, input.Title, input.Content); contentsErr != nil {
			return ErrorResult("UpdateDocumentContents failed: " + contentsErr.Error()), nil, nil
		}
	}

	// Update status
	if input.State != "" {
		if statusErr := datahubClient.UpdateDocumentStatus(ctx, input.URN, input.State); statusErr != nil {
			return ErrorResult("UpdateDocumentStatus failed: " + statusErr.Error()), nil, nil
		}
	}

	// Update related assets
	if input.RelatedAssets != nil {
		if assetsErr := datahubClient.UpdateDocumentRelatedEntities(ctx, input.URN, input.RelatedAssets); assetsErr != nil {
			return ErrorResult("UpdateDocumentRelatedEntities failed: " + assetsErr.Error()), nil, nil
		}
	}

	output := UpdateDocumentOutput{
		URN:    input.URN,
		Action: "updated",
	}

	jsonResult, jsonErr := JSONResult(output)
	if jsonErr != nil {
		return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
	}
	return jsonResult, &output, nil
}
