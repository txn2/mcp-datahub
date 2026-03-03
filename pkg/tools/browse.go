package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BrowseInput is the input for the datahub_browse tool.
type BrowseInput struct {
	//nolint:lll // struct tag cannot be split
	What   string `json:"what" jsonschema_description:"What to browse: tags, domains, or data_products" jsonschema_enum:"tags,domains,data_products"`
	Filter string `json:"filter,omitempty" jsonschema_description:"Optional filter string (tags only)"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerBrowseTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		browseInput, ok := input.(BrowseInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleBrowse(ctx, req, browseInput)
	}

	wrappedHandler := t.wrapHandler(ToolBrowse, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:         string(ToolBrowse),
		Description:  t.getDescription(ToolBrowse, cfg),
		Annotations:  t.getAnnotations(ToolBrowse, cfg),
		Icons:        t.getIcons(ToolBrowse, cfg),
		Title:        t.getTitle(ToolBrowse, cfg),
		OutputSchema: t.getOutputSchema(ToolBrowse, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input BrowseInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleBrowse(ctx context.Context, _ *mcp.CallToolRequest, input BrowseInput) (*mcp.CallToolResult, any, error) {
	if input.What == "" {
		return ErrorResult("what parameter is required (tags, domains, or data_products)"), nil, nil
	}

	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	switch input.What {
	case "tags":
		tags, tagsErr := datahubClient.ListTags(ctx, input.Filter)
		if tagsErr != nil {
			return ErrorResult(tagsErr.Error()), nil, nil
		}
		output := BrowseOutput{Tags: tags}
		jsonResult, jsonErr := JSONResult(output)
		if jsonErr != nil {
			return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
		}
		return jsonResult, &output, nil

	case "domains":
		domains, domainsErr := datahubClient.ListDomains(ctx)
		if domainsErr != nil {
			return ErrorResult(domainsErr.Error()), nil, nil
		}
		output := BrowseOutput{Domains: domains}
		jsonResult, jsonErr := JSONResult(output)
		if jsonErr != nil {
			return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
		}
		return jsonResult, &output, nil

	case "data_products":
		products, productsErr := datahubClient.ListDataProducts(ctx)
		if productsErr != nil {
			return ErrorResult(productsErr.Error()), nil, nil
		}
		output := BrowseOutput{DataProducts: products}
		jsonResult, jsonErr := JSONResult(output)
		if jsonErr != nil {
			return ErrorResult("failed to format result: " + jsonErr.Error()), nil, nil
		}
		return jsonResult, &output, nil

	default:
		return ErrorResult("invalid what value: " + input.What + " (valid: tags, domains, data_products)"), nil, nil
	}
}
