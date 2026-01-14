package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListDomainsInput is the input for the list_domains tool (empty).
type ListDomainsInput struct{}

func (t *Toolkit) registerListDomainsTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		return t.handleListDomains(ctx, req)
	}

	wrappedHandler := t.wrapHandler(ToolListDomains, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolListDomains),
		Description: "List data domains in the DataHub catalog",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDomainsInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleListDomains(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, any, error) {
	domains, err := t.client.ListDomains(ctx)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(domains)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
