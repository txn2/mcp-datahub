package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetGlossaryTermInput is the input for the get_glossary_term tool.
type GetGlossaryTermInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the glossary term"`
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
		Description: "Get a glossary term definition and its related assets",
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

	term, err := t.client.GetGlossaryTerm(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(term)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
