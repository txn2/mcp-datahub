package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DeleteInput is the input for the datahub_delete tool.
type DeleteInput struct {
	//nolint:lll // struct tag cannot be split
	What string `json:"what" jsonschema_description:"Entity type to delete: query, tag, domain, glossary_entity, data_product, application, document, or structured_property" jsonschema_enum:"query,tag,domain,glossary_entity,data_product,application,document,structured_property"`

	URN        string `json:"urn" jsonschema_description:"URN of the entity to delete"`
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerDeleteTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		deleteInput, ok := input.(DeleteInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleDelete(ctx, req, deleteInput)
	}

	wrappedHandler := t.wrapHandler(ToolDelete, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolDelete),
		Description:  t.getDescription(ToolDelete, cfg),
		Annotations:  t.getAnnotations(ToolDelete, cfg),
		Icons:        t.getIcons(ToolDelete, cfg),
		Title:        t.getTitle(ToolDelete, cfg),
		OutputSchema: t.getOutputSchema(ToolDelete, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteInput) (*mcp.CallToolResult, *DeleteOutput, error) {
		result, out, err := wrappedHandler(ctx, req, input)
		if typed, ok := out.(*DeleteOutput); ok {
			return result, typed, err
		}
		return result, nil, err
	})
}

func (t *Toolkit) handleDelete(
	ctx context.Context, _ *mcp.CallToolRequest, input DeleteInput,
) (*mcp.CallToolResult, any, error) {
	if input.What == "" {
		return ErrorResult("what parameter is required"), nil, nil
	}
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	switch input.What {
	case "query":
		err = datahubClient.DeleteQuery(ctx, input.URN)
	case "tag":
		err = datahubClient.DeleteTag(ctx, input.URN)
	case "domain":
		err = datahubClient.DeleteDomain(ctx, input.URN)
	case "glossary_entity":
		err = datahubClient.DeleteGlossaryEntity(ctx, input.URN)
	case "data_product":
		err = datahubClient.DeleteDataProduct(ctx, input.URN)
	case "application":
		err = datahubClient.DeleteApplication(ctx, input.URN)
	case "document":
		err = datahubClient.DeleteDocument(ctx, input.URN)
	case "structured_property":
		err = datahubClient.DeleteStructuredProperty(ctx, input.URN)
	default:
		return ErrorResult("invalid what value: " + input.What), nil, nil
	}

	if err != nil {
		return ErrorResult("Delete " + input.What + " failed: " + err.Error()), nil, nil
	}

	output := DeleteOutput{URN: input.URN, What: input.What, Action: "deleted"}
	jsonResult, jsonErr := JSONResult(output)
	if jsonErr != nil {
		return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
	}
	return jsonResult, &output, nil
}
