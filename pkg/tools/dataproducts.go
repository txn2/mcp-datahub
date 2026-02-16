package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListDataProductsInput is the input for the list_data_products tool.
type ListDataProductsInput struct {
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerListDataProductsTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		productsInput, ok := input.(ListDataProductsInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleListDataProducts(ctx, req, productsInput)
	}

	wrappedHandler := t.wrapHandler(ToolListDataProducts, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolListDataProducts),
		Description: t.getDescription(ToolListDataProducts, cfg),
		Annotations: t.getAnnotations(ToolListDataProducts, cfg),
		Icons:       t.getIcons(ToolListDataProducts, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDataProductsInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleListDataProducts(
	ctx context.Context, _ *mcp.CallToolRequest, input ListDataProductsInput,
) (*mcp.CallToolResult, any, error) {
	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	products, err := datahubClient.ListDataProducts(ctx)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(products)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}

// GetDataProductInput is the input for the get_data_product tool.
type GetDataProductInput struct {
	URN string `json:"urn" jsonschema_description:"The DataHub URN of the data product"`
	// Connection is the named connection to use. Empty uses the default connection.
	Connection string `json:"connection,omitempty" jsonschema_description:"Named connection to use (see datahub_list_connections)"`
}

func (t *Toolkit) registerGetDataProductTool(server *mcp.Server, cfg *toolConfig) {
	baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		productInput, ok := input.(GetDataProductInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return t.handleGetDataProduct(ctx, req, productInput)
	}

	wrappedHandler := t.wrapHandler(ToolGetDataProduct, baseHandler, cfg)

	mcp.AddTool(server, &mcp.Tool{
		Name:        string(ToolGetDataProduct),
		Description: t.getDescription(ToolGetDataProduct, cfg),
		Annotations: t.getAnnotations(ToolGetDataProduct, cfg),
		Icons:       t.getIcons(ToolGetDataProduct, cfg),
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetDataProductInput) (*mcp.CallToolResult, any, error) {
		return wrappedHandler(ctx, req, input)
	})
}

func (t *Toolkit) handleGetDataProduct(
	ctx context.Context, _ *mcp.CallToolRequest, input GetDataProductInput,
) (*mcp.CallToolResult, any, error) {
	if input.URN == "" {
		return ErrorResult("urn parameter is required"), nil, nil
	}

	// Get client for the specified connection
	datahubClient, err := t.getClient(input.Connection)
	if err != nil {
		return ErrorResult("Connection error: " + err.Error()), nil, nil
	}

	product, err := datahubClient.GetDataProduct(ctx, input.URN)
	if err != nil {
		return ErrorResult(err.Error()), nil, nil
	}

	jsonResult, err := JSONResult(product)
	if err != nil {
		return ErrorResult("failed to format result: " + err.Error()), nil, nil
	}

	return jsonResult, nil, nil
}
