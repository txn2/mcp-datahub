package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AddGlossaryTermInput is the input for the add_glossary_term tool.
type AddGlossaryTermInput struct {
	URN        string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	TermURN    string `json:"term_urn" jsonschema_description:"The URN of the glossary term to add (e.g., urn:li:glossaryTerm:Classification)"`
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// RemoveGlossaryTermInput is the input for the remove_glossary_term tool.
type RemoveGlossaryTermInput struct {
	URN        string `json:"urn" jsonschema_description:"The DataHub URN of the entity"`
	TermURN    string `json:"term_urn" jsonschema_description:"The URN of the glossary term to remove"`
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerAddGlossaryTermTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		termInput, ok := input.(AddGlossaryTermInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleAddGlossaryTerm(ctx, req, termInput)
	}

	wrappedHandler := t.wrapHandler(ToolAddGlossaryTerm, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolAddGlossaryTerm),
		Description: "Add a glossary term to a DataHub entity",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddGlossaryTermInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) registerRemoveGlossaryTermTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		termInput, ok := input.(RemoveGlossaryTermInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleRemoveGlossaryTerm(ctx, req, termInput)
	}

	wrappedHandler := t.wrapHandler(ToolRemoveGlossaryTerm, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolRemoveGlossaryTerm),
		Description: "Remove a glossary term from a DataHub entity",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RemoveGlossaryTermInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleAddGlossaryTerm(
	ctx context.Context, _ *mcp.CallToolRequest, input AddGlossaryTermInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}
	if input.TermURN == "" {
		return ErrorResult("term_urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.AddGlossaryTerm(ctx, input.URN, input.TermURN)
	if err != nil {
		return ErrorResult("AddGlossaryTerm failed: " + err.Error()), nil, nil
	}

	result := map[string]string{
		"urn":    input.URN,
		"term":   input.TermURN,
		"aspect": "glossaryTerms",
		"action": "added",
	}

	jsonResult, err := JSONResult(result)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, nil, nil
}

func (t *Toolkit) handleRemoveGlossaryTerm(
	ctx context.Context, _ *mcp.CallToolRequest, input RemoveGlossaryTermInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}
	if input.TermURN == "" {
		return ErrorResult("term_urn parameter is required"), nil, nil
	}

	datahubClient, err := t.getWriteClient(input.Connection)
	if err != nil {
		return ErrorResult("Write error: " + err.Error()), nil, nil
	}

	err = datahubClient.RemoveGlossaryTerm(ctx, input.URN, input.TermURN)
	if err != nil {
		return ErrorResult("RemoveGlossaryTerm failed: " + err.Error()), nil, nil
	}

	result := map[string]string{
		"urn":    input.URN,
		"term":   input.TermURN,
		"aspect": "glossaryTerms",
		"action": "removed",
	}

	jsonResult, err := JSONResult(result)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}
	return jsonResult, nil, nil
}
