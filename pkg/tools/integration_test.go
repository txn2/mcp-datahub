package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/integration"
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

// setupWriteTestServer creates a test MCP server with write tools enabled.
func setupWriteTestServer(t *testing.T, mock *mockClient) *mcp.ClientSession {
	t.Helper()

	toolkit := NewToolkit(mock, Config{WriteEnabled: true})
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

// richQueryProvider is a query provider that returns non-nil values for all methods,
// exercising the with-query-provider code paths in entity, lineage, and schema handlers.
type richQueryProvider struct{}

func (r *richQueryProvider) Name() string { return "rich-mock" }
func (r *richQueryProvider) ResolveTable(_ context.Context, urn string) (*integration.TableIdentifier, error) {
	return &integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: "tbl"}, nil
}
func (r *richQueryProvider) GetTableAvailability(_ context.Context, _ string) (*integration.TableAvailability, error) {
	avail := true
	_ = avail
	return &integration.TableAvailability{Available: true, Connection: "default"}, nil
}
func (r *richQueryProvider) GetQueryExamples(_ context.Context, _ string) ([]integration.QueryExample, error) {
	return []integration.QueryExample{{Name: "sample", SQL: "SELECT 1"}}, nil
}
func (r *richQueryProvider) GetExecutionContext(_ context.Context, _ []string) (*integration.ExecutionContext, error) {
	return &integration.ExecutionContext{
		Source:      "trino",
		Connections: []string{"default"},
		Tables:      map[string]*integration.TableIdentifier{"urn:li:dataset:test": {Catalog: "cat", Schema: "sch", Table: "tbl"}},
	}, nil
}
func (r *richQueryProvider) Close() error { return nil }

// TestToolsViaServer_SchemaValidation exercises the three tools whose outputSchema was
// previously mismatched against their actual return types. It uses rich mock data
// (owners, tags, domain as objects; assets as strings) so that go-sdk's applySchema
// call exercises every type constraint in the schema.  A non-nil CallTool error
// means schema validation failed inside go-sdk.
func TestToolsViaServer_SchemaValidation(t *testing.T) {
	domain := &types.Domain{URN: "urn:li:domain:finance", Name: "Finance"}
	owner := types.Owner{URN: "urn:li:corpuser:alice", Name: "Alice"}

	mock := &mockClient{
		getEntityFunc: func(_ context.Context, urn string) (*types.Entity, error) {
			return &types.Entity{
				URN:    urn,
				Name:   "Rich Entity",
				Type:   "dataset",
				Owners: []types.Owner{owner},
				Tags:   []types.Tag{{URN: "urn:li:tag:PII", Name: "PII"}},
				Domain: domain,
			}, nil
		},
		getLineageFunc: func(_ context.Context, urn string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{
				Start:     urn,
				Direction: "DOWNSTREAM",
				Nodes:     []types.LineageNode{{URN: "urn:li:dataset:up", Name: "upstream", Type: "dataset"}},
			}, nil
		},
		getDataProductFunc: func(_ context.Context, urn string) (*types.DataProduct, error) {
			return &types.DataProduct{
				URN:    urn,
				Name:   "Rich Product",
				Domain: domain,
				Owners: []types.Owner{owner},
				Assets: []string{"urn:li:dataset:a", "urn:li:dataset:b"},
			}, nil
		},
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return &types.SchemaMetadata{Name: "schema", Fields: []types.SchemaField{{FieldPath: "id", Type: "string"}}}, nil
		},
	}

	// Test without query provider (exercises direct-return code paths).
	t.Run("without_query_provider", func(t *testing.T) {
		toolkit := NewToolkit(mock, DefaultConfig())
		impl := &mcp.Implementation{Name: "test-server", Version: "1.0.0"}
		server := mcp.NewServer(impl, nil)
		toolkit.RegisterAll(server)

		serverTransport, clientTransport := mcp.NewInMemoryTransports()
		mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
		go func() { _ = server.Run(context.Background(), serverTransport) }()
		session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
		if err != nil {
			t.Fatalf("connect: %v", err)
		}

		for _, tc := range []struct {
			name string
			tool ToolName
			args map[string]any
		}{
			{"get_entity", ToolGetEntity, map[string]any{"urn": "urn:li:dataset:test"}},
			{"get_lineage", ToolGetLineage, map[string]any{"urn": "urn:li:dataset:test"}},
			{"get_data_product", ToolGetDataProduct, map[string]any{"urn": "urn:li:dataProduct:test"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				result, callErr := session.CallTool(context.Background(), &mcp.CallToolParams{
					Name:      string(tc.tool),
					Arguments: tc.args,
				})
				if callErr != nil {
					t.Errorf("CallTool(%s) schema validation error: %v", tc.name, callErr)
				}
				if result == nil {
					t.Errorf("CallTool(%s) returned nil result", tc.name)
				}
			})
		}
	})

	// Test with query provider (exercises enrichment + query_table stringify code paths).
	t.Run("with_query_provider", func(t *testing.T) {
		toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(&richQueryProvider{}))
		impl := &mcp.Implementation{Name: "test-server", Version: "1.0.0"}
		server := mcp.NewServer(impl, nil)
		toolkit.RegisterAll(server)

		serverTransport, clientTransport := mcp.NewInMemoryTransports()
		mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
		go func() { _ = server.Run(context.Background(), serverTransport) }()
		session, err := mcpClient.Connect(context.Background(), clientTransport, nil)
		if err != nil {
			t.Fatalf("connect: %v", err)
		}

		for _, tc := range []struct {
			name string
			tool ToolName
			args map[string]any
		}{
			{"get_entity_enriched", ToolGetEntity, map[string]any{"urn": "urn:li:dataset:test"}},
			{"get_lineage_enriched", ToolGetLineage, map[string]any{"urn": "urn:li:dataset:test"}},
			{"get_data_product_enriched", ToolGetDataProduct, map[string]any{"urn": "urn:li:dataProduct:test"}},
		} {
			t.Run(tc.name, func(t *testing.T) {
				result, callErr := session.CallTool(context.Background(), &mcp.CallToolParams{
					Name:      string(tc.tool),
					Arguments: tc.args,
				})
				if callErr != nil {
					t.Errorf("CallTool(%s) schema validation error: %v", tc.name, callErr)
				}
				if result == nil {
					t.Errorf("CallTool(%s) returned nil result", tc.name)
				}
			})
		}
	})
}

// TestWriteToolsViaServer tests write tools through the MCP server.
// This covers the register*Tool closures including type assertions.
func TestWriteToolsViaServer(t *testing.T) {
	mock := createTestMockClient()
	session := setupWriteTestServer(t, mock)

	tests := []struct {
		name      string
		toolName  ToolName
		arguments map[string]any
	}{
		{
			"update_description", ToolUpdateDescription,
			map[string]any{
				"urn":         "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"description": "Updated description",
			},
		},
		{
			"add_tag", ToolAddTag,
			map[string]any{
				"urn":     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"tag_urn": "urn:li:tag:PII",
			},
		},
		{
			"remove_tag", ToolRemoveTag,
			map[string]any{
				"urn":     "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"tag_urn": "urn:li:tag:PII",
			},
		},
		{
			"add_glossary_term", ToolAddGlossaryTerm,
			map[string]any{
				"urn":      "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"term_urn": "urn:li:glossaryTerm:Classification",
			},
		},
		{
			"remove_glossary_term", ToolRemoveGlossaryTerm,
			map[string]any{
				"urn":      "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"term_urn": "urn:li:glossaryTerm:Classification",
			},
		},
		{
			"add_link", ToolAddLink,
			map[string]any{
				"urn":         "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"url":         "https://docs.example.com",
				"description": "Documentation",
			},
		},
		{
			"remove_link", ToolRemoveLink,
			map[string]any{
				"urn": "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)",
				"url": "https://docs.example.com",
			},
		},
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
