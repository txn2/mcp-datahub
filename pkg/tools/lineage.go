package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
)

// GetLineageInput is the input for the get_lineage tool.
type GetLineageInput struct {
	URN       string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	Direction string `json:"direction,omitempty" jsonschema_description:"Lineage direction: UPSTREAM or DOWNSTREAM (default: DOWNSTREAM)"`
	Depth     int    `json:"depth,omitempty" jsonschema_description:"Maximum depth of lineage traversal (default: 1, max: 5)"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetLineageTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		lineageInput, ok := input.(GetLineageInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetLineage(ctx, req, lineageInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetLineage, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolGetLineage),
		Description: "Get upstream or downstream lineage for an entity",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetLineageInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetLineage(ctx context.Context, _ *mcp.CallToolRequest, input GetLineageInput) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	var opts []client.LineageOption
	if input.Direction != "" {
		opts = append(opts, client.WithDirection(input.Direction))
	}
	if input.Depth > 0 {
		opts = append(opts, client.WithDepth(input.Depth))
	}

	lineage, err := datahubClient.GetLineage(ctx, input.URN, opts...)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(lineage)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
