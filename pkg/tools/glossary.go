package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetGlossaryTermInput is the input for the get_glossary_term tool.
type GetGlossaryTermInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the glossary term"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetGlossaryTermTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		glossaryInput, ok := input.(GetGlossaryTermInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetGlossaryTerm(ctx, req, glossaryInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetGlossaryTerm, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolGetGlossaryTerm),
		Description: t.getDescription(ToolGetGlossaryTerm, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetGlossaryTermInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetGlossaryTerm(
	ctx context.Context, _ *mcp.CallToolRequest, input GetGlossaryTermInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	term, err := datahubClient.GetGlossaryTerm(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(term)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
