package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// createTestMockClient returns a mock client with all methods stubbed.
func createTestMockClient() *mockClient {
	return &mockClient{
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
}

// setupTestServer creates and starts a test MCP server with the toolkit registered.
func setupTestServer(t *testing.T, mock *mockClient) *mcp.ClientSession {
	t.Helper()

	toolkit := NewToolkit(mock, DefaultConfig())
	impl := &mcp.Implementation{Name: "test-server", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.RegisterAll(server)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)

	go func() {
		_ = server.Run(context.Background(), serverTransport)
	}()

	session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
	if err != nil {
		t.Fatalf("Failed to connect client: %v", err)
	}

	return session
}

// TestToolsViaServer tests tools by actually invoking them through the MCP server.
// This covers the registration code paths including the type assertion branches.
func TestToolsViaServer(t *testing.T) {
	mock := createTestMockClient()
	session := setupTestServer(t, mock)

	tests := []struct {
		name      string
		toolName  ToolName
		arguments map[string]any
	}{
		{"search", ToolSearch, map[string]any{"query": "test"}},
		{"get_entity", ToolGetEntity, map[string]any{"urn": "urn:li:dataset:test"}},
		{"list_domains", ToolListDomains, map[string]any{}},
		{"list_tags", ToolListTags, map[string]any{}},
		{"get_schema", ToolGetSchema, map[string]any{"urn": "urn:li:dataset:test"}},
		{"get_lineage", ToolGetLineage, map[string]any{"urn": "urn:li:dataset:test"}},
		{"get_queries", ToolGetQueries, map[string]any{"urn": "urn:li:dataset:test"}},
		{"get_glossary_term", ToolGetGlossaryTerm, map[string]any{"urn": "urn:li:glossaryTerm:test"}},
		{"list_data_products", ToolListDataProducts, map[string]any{}},
		{"get_data_product", ToolGetDataProduct, map[string]any{"urn": "urn:li:dataProduct:test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
				Name:      string(tt.toolName),
				Arguments: tt.arguments,
			})
			if err != nil {
				t.Errorf("CallTool(%s) error: %v", tt.name, err)
			}
			if result == nil {
				t.Errorf("CallTool(%s) returned nil result", tt.name)
			}
		})
	}
}
