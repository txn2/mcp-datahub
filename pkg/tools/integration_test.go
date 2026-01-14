package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// TestToolsViaServer tests tools by actually invoking them through the MCP server.
// This covers the registration code paths including the type assertion branches.
func TestToolsViaServer(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{Total: 1, Entities: []types.SearchEntity{{URN: "urn:li:dataset:test"}}}, nil
		},
		getEntityFunc: func(_ context.Context, urn string) (*types.Entity, error) {
			return &types.Entity{URN: urn, Name: "test"}, nil
		},
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return &types.SchemaMetadata{Name: "schema"}, nil
		},
		getLineageFunc: func(_ context.Context, urn string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{Start: urn}, nil
		},
		getQueriesFunc: func(_ context.Context, _ string) (*types.QueryList, error) {
			return &types.QueryList{Total: 0}, nil
		},
		getGlossaryTermFunc: func(_ context.Context, urn string) (*types.GlossaryTerm, error) {
			return &types.GlossaryTerm{URN: urn}, nil
		},
		listTagsFunc: func(_ context.Context, _ string) ([]types.Tag, error) {
			return []types.Tag{{URN: "urn:li:tag:test", Name: "test"}}, nil
		},
		listDomainsFunc: func(_ context.Context) ([]types.Domain, error) {
			return []types.Domain{{URN: "urn:li:domain:test", Name: "test"}}, nil
		},
		listDataProductsFunc: func(_ context.Context) ([]types.DataProduct, error) {
			return []types.DataProduct{{URN: "urn:li:dataProduct:test", Name: "test"}}, nil
		},
		getDataProductFunc: func(_ context.Context, urn string) (*types.DataProduct, error) {
			return &types.DataProduct{URN: urn, Name: "test"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())

	impl := &mcp.Implementation{Name: "test-server", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)

	// Register all tools
	toolkit.RegisterAll(server)

	// Create in-memory transports
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	// Create client
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)

	// Run server in background
	go func() {
		_ = server.Run(context.Background(), serverTransport)
	}()

	// Connect client
	session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}

	// Test calling search tool
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolSearch),
		Arguments: map[string]any{"query": "test"},
	})
	if err != nil {
		t.Errorf("CallTool(search) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(search) returned nil result")
	}

	// Test calling get_entity tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolGetEntity),
		Arguments: map[string]any{"urn": "urn:li:dataset:test"},
	})
	if err != nil {
		t.Errorf("CallTool(get_entity) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(get_entity) returned nil result")
	}

	// Test calling list_domains tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolListDomains),
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Errorf("CallTool(list_domains) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(list_domains) returned nil result")
	}

	// Test calling list_tags tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolListTags),
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Errorf("CallTool(list_tags) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(list_tags) returned nil result")
	}

	// Test calling get_schema tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolGetSchema),
		Arguments: map[string]any{"urn": "urn:li:dataset:test"},
	})
	if err != nil {
		t.Errorf("CallTool(get_schema) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(get_schema) returned nil result")
	}

	// Test calling get_lineage tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolGetLineage),
		Arguments: map[string]any{"urn": "urn:li:dataset:test"},
	})
	if err != nil {
		t.Errorf("CallTool(get_lineage) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(get_lineage) returned nil result")
	}

	// Test calling get_queries tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolGetQueries),
		Arguments: map[string]any{"urn": "urn:li:dataset:test"},
	})
	if err != nil {
		t.Errorf("CallTool(get_queries) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(get_queries) returned nil result")
	}

	// Test calling get_glossary_term tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolGetGlossaryTerm),
		Arguments: map[string]any{"urn": "urn:li:glossaryTerm:test"},
	})
	if err != nil {
		t.Errorf("CallTool(get_glossary_term) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(get_glossary_term) returned nil result")
	}

	// Test calling list_data_products tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolListDataProducts),
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Errorf("CallTool(list_data_products) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(list_data_products) returned nil result")
	}

	// Test calling get_data_product tool
	result, err = session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      string(ToolGetDataProduct),
		Arguments: map[string]any{"urn": "urn:li:dataProduct:test"},
	})
	if err != nil {
		t.Errorf("CallTool(get_data_product) error: %v", err)
	}
	if result == nil {
		t.Error("CallTool(get_data_product) returned nil result")
	}
}
