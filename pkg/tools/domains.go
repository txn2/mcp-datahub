package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListDomainsInput is the input for the deprecated datahub_list_domains tool.
//
// Deprecated: use BrowseInput with What="domains" instead.
type ListDomainsInput struct {
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

// registerListDomainsTool registers the deprecated datahub_list_domains alias.
// It delegates to handleBrowse with What="domains".
func (t *Toolkit) registerListDomainsTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		domainsInput, ok := input.(ListDomainsInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		browseInput := BrowseInput{
			What:       "domains",
			Connection: domainsInput.Connection,
		}
		return t.handleBrowse(ctx, req, browseInput)
	}

	wrappedHandler := t.wrapHandler(ToolListDomains, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolListDomains),
		Description:  t.getDescription(ToolListDomains, cfg),
		Annotations:  t.getAnnotations(ToolListDomains, cfg),
		Icons:        t.getIcons(ToolListDomains, cfg),
		Title:        t.getTitle(ToolListDomains, cfg),
		OutputSchema: t.getOutputSchema(ToolListDomains, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDomainsInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}
