package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetColumnLineageInput is the input for the get_column_lineage tool.
type GetColumnLineageInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the dataset"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetColumnLineageTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		colLineageInput, ok := input.(GetColumnLineageInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetColumnLineage(ctx, req, colLineageInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetColumnLineage, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolGetColumnLineage),
		Description: t.getDescription(ToolGetColumnLineage, cfg),
		Annotations: t.getAnnotations(ToolGetColumnLineage, cfg),
		Icons:       t.getIcons(ToolGetColumnLineage, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetColumnLineageInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetColumnLineage(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input GetColumnLineageInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	columnLineage, err := datahubClient.GetColumnLineage(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(columnLineage)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
