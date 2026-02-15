package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListDomainsInput is the input for the list_domains tool.
type ListDomainsInput struct {
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerListDomainsTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		domainsInput, ok := input.(ListDomainsInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleListDomains(ctx, req, domainsInput)
	}

	wrappedHandler := t.wrapHandler(ToolListDomains, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolListDomains),
		Description: t.getDescription(ToolListDomains, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDomainsInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleListDomains(ctx context.Context, _ *mcp.CallToolRequest, input ListDomainsInput) (*mcp.CallToolResult, any, error) {
	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	domains, err := datahubClient.ListDomains(ctx)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(domains)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
