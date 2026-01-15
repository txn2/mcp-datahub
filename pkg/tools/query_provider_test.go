package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/integration"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// fullMockQueryProvider implements QueryProvider with configurable behavior.
type fullMockQueryProvider struct {
	name                   string
	resolveTableFn         func(ctx context.Context, urn string) (*integration.TableIdentifier, error)
	getTableAvailabilityFn func(ctx context.Context, urn string) (*integration.TableAvailability, error)
	getQueryExamplesFn     func(ctx context.Context, urn string) ([]integration.QueryExample, error)
	getExecutionContextFn  func(ctx context.Context, urns []string) (*integration.ExecutionContext, error)
}

func (m *fullMockQueryProvider) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mock"
}

func (m *fullMockQueryProvider) ResolveTable(ctx context.Context, urn string) (*integration.TableIdentifier, error) {
	if m.resolveTableFn != nil {
		return m.resolveTableFn(ctx, urn)
	}
	return nil, nil
}

func (m *fullMockQueryProvider) GetTableAvailability(
	ctx context.Context,
	urn string,
) (*integration.TableAvailability, error) {
	if m.getTableAvailabilityFn != nil {
		return m.getTableAvailabilityFn(ctx, urn)
	}
	return nil, nil
}

func (m *fullMockQueryProvider) GetQueryExamples(
	ctx context.Context,
	urn string,
) ([]integration.QueryExample, error) {
	if m.getQueryExamplesFn != nil {
		return m.getQueryExamplesFn(ctx, urn)
	}
	return nil, nil
}

func (m *fullMockQueryProvider) GetExecutionContext(
	ctx context.Context,
	urns []string,
) (*integration.ExecutionContext, error) {
	if m.getExecutionContextFn != nil {
		return m.getExecutionContextFn(ctx, urns)
	}
	return nil, nil
}

func (m *fullMockQueryProvider) Close() error { return nil }

// Helper to extract text from result.
func extractResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

// Tests for QueryProvider integration in search handler.

func TestHandleSearch_WithQueryProvider(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Entities: []types.SearchEntity{
					{URN: "urn:li:dataset:1", Name: "Table1"},
					{URN: "urn:li:dataset:2", Name: "Table2"},
				},
				Total: 2,
			}, nil
		},
	}

	provider := &fullMockQueryProvider{
		getTableAvailabilityFn: func(_ context.Context, urn string) (*integration.TableAvailability, error) {
			if urn == "urn:li:dataset:1" {
				return &integration.TableAvailability{
					Available: true,
					Table:     &integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: "t1"},
				}, nil
			}
			return &integration.TableAvailability{Available: false}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if !strings.Contains(text, "query_context") {
		t.Error("expected query_context in result")
	}
	if !strings.Contains(text, "cat.sch.t1") {
		t.Error("expected table identifier in result")
	}
}

func TestHandleSearch_WithQueryProvider_NoResults(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{Entities: []types.SearchEntity{}, Total: 0}, nil
		},
	}

	called := false
	provider := &fullMockQueryProvider{
		getTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			called = true
			return nil, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	_, _, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Error("provider should not be called for empty results")
	}
}

func TestHandleSearch_WithQueryProvider_Error(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Entities: []types.SearchEntity{{URN: "urn:li:dataset:1"}},
				Total:    1,
			}, nil
		},
	}

	provider := &fullMockQueryProvider{
		getTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			return nil, errors.New("provider error")
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still succeed but without query context
	if result.IsError {
		t.Error("should succeed even with provider error")
	}
}

func TestHandleSearch_WithoutQueryProvider(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Entities: []types.SearchEntity{{URN: "urn:li:dataset:1"}},
				Total:    1,
			}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if strings.Contains(text, "query_context") {
		t.Error("should not have query_context without provider")
	}
}

// Tests for QueryProvider integration in entity handler.

func TestHandleGetEntity_WithQueryProvider(t *testing.T) {
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return &types.Entity{URN: "urn:li:dataset:test", Name: "Test"}, nil
		},
	}

	provider := &fullMockQueryProvider{
		resolveTableFn: func(_ context.Context, _ string) (*integration.TableIdentifier, error) {
			return &integration.TableIdentifier{Catalog: "hive", Schema: "db", Table: "test"}, nil
		},
		getQueryExamplesFn: func(_ context.Context, _ string) ([]integration.QueryExample, error) {
			return []integration.QueryExample{
				{Name: "sample", SQL: "SELECT * FROM hive.db.test LIMIT 10", Category: "sample"},
			}, nil
		},
		getTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			return &integration.TableAvailability{Available: true}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetEntity(
		context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if !strings.Contains(text, "query_table") {
		t.Error("expected query_table in result")
	}
	if !strings.Contains(text, "query_examples") {
		t.Error("expected query_examples in result")
	}
	if !strings.Contains(text, "query_availability") {
		t.Error("expected query_availability in result")
	}
	if !strings.Contains(text, "hive.db.test") {
		t.Error("expected table identifier in result")
	}
}

func TestHandleGetEntity_WithQueryProvider_NilResults(t *testing.T) {
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return &types.Entity{URN: "urn:li:dataset:test"}, nil
		},
	}

	provider := &fullMockQueryProvider{
		resolveTableFn: func(_ context.Context, _ string) (*integration.TableIdentifier, error) {
			return nil, nil
		},
		getQueryExamplesFn: func(_ context.Context, _ string) ([]integration.QueryExample, error) {
			return nil, nil
		},
		getTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			return nil, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetEntity(
		context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	// Should still have entity but no query fields
	text := extractResultText(t, result)
	if !strings.Contains(text, "entity") {
		t.Error("expected entity in result")
	}
}

func TestHandleGetEntity_WithoutQueryProvider(t *testing.T) {
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return &types.Entity{URN: "urn:li:dataset:test"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, err := toolkit.handleGetEntity(
		context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if strings.Contains(text, "query_table") {
		t.Error("should not have query_table without provider")
	}
}

func TestHandleGetEntity_WithQueryProvider_Errors(t *testing.T) {
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return &types.Entity{URN: "urn:li:dataset:test"}, nil
		},
	}

	provider := &fullMockQueryProvider{
		resolveTableFn: func(_ context.Context, _ string) (*integration.TableIdentifier, error) {
			return nil, errors.New("resolve error")
		},
		getQueryExamplesFn: func(_ context.Context, _ string) ([]integration.QueryExample, error) {
			return nil, errors.New("examples error")
		},
		getTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			return nil, errors.New("availability error")
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetEntity(
		context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still succeed even with provider errors
	if result.IsError {
		t.Error("should succeed even with provider errors")
	}
}

// Tests for QueryProvider integration in schema handler.

func TestHandleGetSchema_WithQueryProvider(t *testing.T) {
	mock := &mockClient{
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return &types.SchemaMetadata{
				Fields: []types.SchemaField{{FieldPath: "id", Type: "INT"}},
			}, nil
		},
	}

	provider := &fullMockQueryProvider{
		resolveTableFn: func(_ context.Context, _ string) (*integration.TableIdentifier, error) {
			return &integration.TableIdentifier{Catalog: "hive", Schema: "db", Table: "test"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetSchema(
		context.Background(), nil, GetSchemaInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if !strings.Contains(text, "schema") {
		t.Error("expected schema in result")
	}
	if !strings.Contains(text, "query_table") {
		t.Error("expected query_table in result")
	}
}

func TestHandleGetSchema_WithoutQueryProvider(t *testing.T) {
	mock := &mockClient{
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return &types.SchemaMetadata{}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, err := toolkit.handleGetSchema(
		context.Background(), nil, GetSchemaInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if strings.Contains(text, "query_table") {
		t.Error("should not have query_table without provider")
	}
}

func TestHandleGetSchema_WithQueryProvider_Error(t *testing.T) {
	mock := &mockClient{
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return &types.SchemaMetadata{}, nil
		},
	}

	provider := &fullMockQueryProvider{
		resolveTableFn: func(_ context.Context, _ string) (*integration.TableIdentifier, error) {
			return nil, errors.New("resolve error")
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetSchema(
		context.Background(), nil, GetSchemaInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still succeed even with provider error
	if result.IsError {
		t.Error("should succeed even with provider error")
	}
}

// Tests for QueryProvider integration in lineage handler.

func TestHandleGetLineage_WithQueryProvider(t *testing.T) {
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{
				Start: "urn:li:dataset:test",
				Nodes: []types.LineageNode{
					{URN: "urn:li:dataset:upstream"},
				},
			}, nil
		},
	}

	provider := &fullMockQueryProvider{
		getExecutionContextFn: func(_ context.Context, urns []string) (*integration.ExecutionContext, error) {
			tables := make(map[string]*integration.TableIdentifier)
			for _, urn := range urns {
				tables[urn] = &integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: urn}
			}
			return &integration.ExecutionContext{Tables: tables, Source: "trino"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetLineage(
		context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if !strings.Contains(text, "lineage") {
		t.Error("expected lineage in result")
	}
	if !strings.Contains(text, "execution_context") {
		t.Error("expected execution_context in result")
	}
}

func TestHandleGetLineage_WithoutQueryProvider(t *testing.T) {
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{Start: "urn:li:dataset:test"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, err := toolkit.handleGetLineage(
		context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatal("expected success result")
	}

	text := extractResultText(t, result)
	if strings.Contains(text, "execution_context") {
		t.Error("should not have execution_context without provider")
	}
}

func TestHandleGetLineage_WithQueryProvider_Error(t *testing.T) {
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{
				Start: "urn:li:dataset:test",
				Nodes: []types.LineageNode{{URN: "urn:li:dataset:node1"}},
			}, nil
		},
	}

	provider := &fullMockQueryProvider{
		getExecutionContextFn: func(_ context.Context, _ []string) (*integration.ExecutionContext, error) {
			return nil, errors.New("context error")
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetLineage(
		context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still succeed even with provider error
	if result.IsError {
		t.Error("should succeed even with provider error")
	}
}

func TestHandleGetLineage_WithQueryProvider_EmptyLineage(t *testing.T) {
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{Start: "urn:li:dataset:test", Nodes: nil}, nil
		},
	}

	called := false
	provider := &fullMockQueryProvider{
		getExecutionContextFn: func(_ context.Context, urns []string) (*integration.ExecutionContext, error) {
			called = true
			// Should only be called with the start URN
			if len(urns) != 1 {
				t.Errorf("expected 1 URN, got %d", len(urns))
			}
			return &integration.ExecutionContext{Source: "trino"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	result, _, err := toolkit.handleGetLineage(
		context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Error("provider should still be called for start URN")
	}

	if result.IsError {
		t.Fatal("expected success result")
	}
}

// Tests for collectLineageURNs.

func TestCollectLineageURNs_MapInput(t *testing.T) {
	lineage := map[string]any{
		"start": "urn:li:dataset:start",
		"nodes": []any{
			map[string]any{"urn": "urn:li:dataset:node1"},
			map[string]any{"urn": "urn:li:dataset:node2"},
		},
	}

	urns := collectLineageURNs(lineage)

	if len(urns) != 3 {
		t.Errorf("expected 3 URNs, got %d", len(urns))
	}

	expected := map[string]bool{
		"urn:li:dataset:start": true,
		"urn:li:dataset:node1": true,
		"urn:li:dataset:node2": true,
	}
	for _, urn := range urns {
		if !expected[urn] {
			t.Errorf("unexpected URN: %s", urn)
		}
	}
}

func TestCollectLineageURNs_StructInput(t *testing.T) {
	lineage := &types.LineageResult{
		Start: "urn:li:dataset:start",
		Nodes: []types.LineageNode{
			{URN: "urn:li:dataset:node1"},
		},
	}

	urns := collectLineageURNs(lineage)

	// Should handle struct via JSON roundtrip
	if len(urns) != 2 {
		t.Errorf("expected 2 URNs, got %d: %v", len(urns), urns)
	}
}

func TestCollectLineageURNs_EmptyInput(t *testing.T) {
	urns := collectLineageURNs(map[string]any{})
	if len(urns) != 0 {
		t.Errorf("expected 0 URNs, got %d", len(urns))
	}
}

func TestCollectLineageURNs_NoNodes(t *testing.T) {
	lineage := map[string]any{
		"start": "urn:li:dataset:start",
	}

	urns := collectLineageURNs(lineage)

	if len(urns) != 1 {
		t.Errorf("expected 1 URN, got %d", len(urns))
	}
}

func TestCollectLineageURNs_NilInput(t *testing.T) {
	urns := collectLineageURNs(nil)
	if len(urns) != 0 {
		t.Errorf("expected 0 URNs, got %d", len(urns))
	}
}

func TestCollectLineageURNs_NonMapInput(t *testing.T) {
	// String input - should return empty
	urns := collectLineageURNs("not a map")
	if len(urns) != 0 {
		t.Errorf("expected 0 URNs for string input, got %d", len(urns))
	}
}

// Tests for NoOpQueryProvider.

func TestNoOpQueryProvider(t *testing.T) {
	p := &integration.NoOpQueryProvider{}

	if p.Name() != "noop" {
		t.Errorf("Name() = %s, want noop", p.Name())
	}

	table, err := p.ResolveTable(context.Background(), "urn")
	if err != nil || table != nil {
		t.Error("ResolveTable should return nil, nil")
	}

	avail, err := p.GetTableAvailability(context.Background(), "urn")
	if err != nil || avail != nil {
		t.Error("GetTableAvailability should return nil, nil")
	}

	examples, err := p.GetQueryExamples(context.Background(), "urn")
	if err != nil || examples != nil {
		t.Error("GetQueryExamples should return nil, nil")
	}

	ctx, err := p.GetExecutionContext(context.Background(), []string{"urn"})
	if err != nil || ctx != nil {
		t.Error("GetExecutionContext should return nil, nil")
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

// Tests for QueryProviderFunc.

func TestQueryProviderFunc_AllMethods(t *testing.T) {
	p := &integration.QueryProviderFunc{
		NameFn: func() string { return "custom" },
		ResolveTableFn: func(_ context.Context, urn string) (*integration.TableIdentifier, error) {
			return &integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: urn}, nil
		},
		GetTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			return &integration.TableAvailability{Available: true}, nil
		},
		GetQueryExamplesFn: func(_ context.Context, _ string) ([]integration.QueryExample, error) {
			return []integration.QueryExample{{Name: "test"}}, nil
		},
		GetExecutionContextFn: func(_ context.Context, _ []string) (*integration.ExecutionContext, error) {
			return &integration.ExecutionContext{Source: "test"}, nil
		},
		CloseFn: func() error { return nil },
	}

	if p.Name() != "custom" {
		t.Errorf("Name() = %s, want custom", p.Name())
	}

	table, err := p.ResolveTable(context.Background(), "test")
	if err != nil || table == nil || table.Table != "test" {
		t.Error("ResolveTable failed")
	}

	avail, err := p.GetTableAvailability(context.Background(), "test")
	if err != nil || avail == nil || !avail.Available {
		t.Error("GetTableAvailability failed")
	}

	examples, err := p.GetQueryExamples(context.Background(), "test")
	if err != nil || len(examples) != 1 {
		t.Error("GetQueryExamples failed")
	}

	execCtx, err := p.GetExecutionContext(context.Background(), []string{"test"})
	if err != nil || execCtx == nil || execCtx.Source != "test" {
		t.Error("GetExecutionContext failed")
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

func TestQueryProviderFunc_NilFunctions(t *testing.T) {
	p := &integration.QueryProviderFunc{}

	if p.Name() != "func" {
		t.Errorf("Name() = %s, want func", p.Name())
	}

	table, err := p.ResolveTable(context.Background(), "test")
	if err != nil || table != nil {
		t.Error("ResolveTable should return nil, nil for nil function")
	}

	avail, err := p.GetTableAvailability(context.Background(), "test")
	if err != nil || avail != nil {
		t.Error("GetTableAvailability should return nil, nil for nil function")
	}

	examples, err := p.GetQueryExamples(context.Background(), "test")
	if err != nil || examples != nil {
		t.Error("GetQueryExamples should return nil, nil for nil function")
	}

	execCtx, err := p.GetExecutionContext(context.Background(), []string{"test"})
	if err != nil || execCtx != nil {
		t.Error("GetExecutionContext should return nil, nil for nil function")
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}

// Tests for TableIdentifier.String.

func TestTableIdentifier_String(t *testing.T) {
	tests := []struct {
		name string
		ti   integration.TableIdentifier
		want string
	}{
		{
			name: "basic",
			ti:   integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: "tbl"},
			want: "cat.sch.tbl",
		},
		{
			name: "with_connection",
			ti: integration.TableIdentifier{
				Connection: "conn", Catalog: "cat", Schema: "sch", Table: "tbl",
			},
			want: "conn:cat.sch.tbl",
		},
		{
			name: "empty_connection",
			ti: integration.TableIdentifier{
				Connection: "", Catalog: "hive", Schema: "default", Table: "users",
			},
			want: "hive.default.users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ti.String(); got != tt.want {
				t.Errorf("TableIdentifier.String() = %s, want %s", got, tt.want)
			}
		})
	}
}

// Tests for option wiring.

func TestWithQueryProvider(t *testing.T) {
	mock := &mockClient{}
	provider := &fullMockQueryProvider{name: "test-provider"}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))

	if toolkit.queryProvider == nil {
		t.Fatal("queryProvider should be set")
	}
	if toolkit.queryProvider.Name() != "test-provider" {
		t.Errorf("queryProvider.Name() = %s, want test-provider", toolkit.queryProvider.Name())
	}
}

func TestWithQueryProvider_Nil(t *testing.T) {
	mock := &mockClient{}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(nil))

	if toolkit.queryProvider != nil {
		t.Error("queryProvider should be nil when nil is passed")
	}
}

// Tests for integration middleware building.

func TestBuildIntegrationMiddleware_AllComponents(t *testing.T) {
	mock := &mockClient{}

	urnResolver := &mockURNResolver{}
	accessFilter := &mockAccessFilter{}
	auditLogger := &mockAuditLogger{}
	enricher := &mockMetadataEnricher{}
	getUserID := func(_ context.Context) string { return "test-user" }

	toolkit := NewToolkit(mock, DefaultConfig(),
		WithURNResolver(urnResolver),
		WithAccessFilter(accessFilter),
		WithAuditLogger(auditLogger, getUserID),
		WithMetadataEnricher(enricher),
	)

	// Should have 4 middleware components
	if len(toolkit.integrationMiddleware) != 4 {
		t.Errorf("expected 4 integration middleware, got %d", len(toolkit.integrationMiddleware))
	}
}

func TestBuildIntegrationMiddleware_PartialComponents(t *testing.T) {
	mock := &mockClient{}

	toolkit := NewToolkit(mock, DefaultConfig(),
		WithAccessFilter(&mockAccessFilter{}),
	)

	// Should have 1 middleware component
	if len(toolkit.integrationMiddleware) != 1 {
		t.Errorf("expected 1 integration middleware, got %d", len(toolkit.integrationMiddleware))
	}
}

func TestBuildIntegrationMiddleware_NoComponents(t *testing.T) {
	mock := &mockClient{}

	toolkit := NewToolkit(mock, DefaultConfig())

	// Should have 0 middleware components
	if len(toolkit.integrationMiddleware) != 0 {
		t.Errorf("expected 0 integration middleware, got %d", len(toolkit.integrationMiddleware))
	}
}
