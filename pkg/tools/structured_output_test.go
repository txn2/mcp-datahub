package tools

import (
	"context"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/integration"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// TestStructuredOutput verifies that all read handlers return a non-nil
// structured output value (second return value) on success, which is required
// for go-sdk to populate structuredContent in the tools/call response.
//
// When a tool declares outputSchema in tools/list, MCP hosts expect
// structuredContent in tools/call responses. go-sdk only populates
// structuredContent when the handler returns a non-nil second value.
func TestStructuredOutput_ListDomains(t *testing.T) {
	domains := []types.Domain{
		{URN: "urn:li:domain:test", Name: "Test"},
	}
	mock := &mockClient{
		listDomainsFunc: func(_ context.Context) ([]types.Domain, error) {
			return domains, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleListDomains(context.Background(), nil, ListDomainsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleListDomains structured output is nil, want non-nil")
	}
	output, ok := out.(*ListDomainsOutput)
	if !ok {
		t.Fatalf("structured output type = %T, want *ListDomainsOutput", out)
	}
	if len(output.Domains) != 1 || output.Domains[0].URN != "urn:li:domain:test" {
		t.Errorf("structured output domains = %v, want 1 domain with correct URN", output.Domains)
	}
}

func TestStructuredOutput_ListTags(t *testing.T) {
	tags := []types.Tag{
		{URN: "urn:li:tag:PII", Name: "PII"},
	}
	mock := &mockClient{
		listTagsFunc: func(_ context.Context, _ string) ([]types.Tag, error) {
			return tags, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleListTags(context.Background(), nil, ListTagsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleListTags structured output is nil, want non-nil")
	}
	output, ok := out.(*ListTagsOutput)
	if !ok {
		t.Fatalf("structured output type = %T, want *ListTagsOutput", out)
	}
	if len(output.Tags) != 1 || output.Tags[0].Name != "PII" {
		t.Errorf("structured output tags = %v, want 1 tag", output.Tags)
	}
}

func TestStructuredOutput_ListDataProducts(t *testing.T) {
	products := []types.DataProduct{
		{URN: "urn:li:dataProduct:test", Name: "Test Product"},
	}
	mock := &mockClient{
		listDataProductsFunc: func(_ context.Context) ([]types.DataProduct, error) {
			return products, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleListDataProducts(context.Background(), nil, ListDataProductsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleListDataProducts structured output is nil, want non-nil")
	}
	output, ok := out.(*ListDataProductsOutput)
	if !ok {
		t.Fatalf("structured output type = %T, want *ListDataProductsOutput", out)
	}
	if len(output.DataProducts) != 1 || output.DataProducts[0].URN != "urn:li:dataProduct:test" {
		t.Errorf("structured output data_products = %v, want 1 product", output.DataProducts)
	}
}

func TestStructuredOutput_GetDataProduct(t *testing.T) {
	product := &types.DataProduct{URN: "urn:li:dataProduct:test", Name: "Test Product"}
	mock := &mockClient{
		getDataProductFunc: func(_ context.Context, _ string) (*types.DataProduct, error) {
			return product, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetDataProduct(context.Background(), nil, GetDataProductInput{URN: "urn:li:dataProduct:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetDataProduct structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetColumnLineage(t *testing.T) {
	lineage := &types.ColumnLineage{
		DatasetURN: "urn:li:dataset:test",
		Mappings:   []types.ColumnLineageMapping{{DownstreamColumn: "col1"}},
	}
	mock := &mockClient{
		getColumnLineageFunc: func(_ context.Context, _ string) (*types.ColumnLineage, error) {
			return lineage, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetColumnLineage(context.Background(), nil, GetColumnLineageInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetColumnLineage structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetQueries(t *testing.T) {
	queries := &types.QueryList{
		Total:   1,
		Queries: []types.Query{{Statement: "SELECT 1"}},
	}
	mock := &mockClient{
		getQueriesFunc: func(_ context.Context, _ string) (*types.QueryList, error) {
			return queries, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetQueries(context.Background(), nil, GetQueriesInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetQueries structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetGlossaryTerm(t *testing.T) {
	term := &types.GlossaryTerm{URN: "urn:li:glossaryTerm:Revenue", Name: "Revenue"}
	mock := &mockClient{
		getGlossaryTermFunc: func(_ context.Context, _ string) (*types.GlossaryTerm, error) {
			return term, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetGlossaryTerm(context.Background(), nil, GetGlossaryTermInput{URN: "urn:li:glossaryTerm:Revenue"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetGlossaryTerm structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetEntity_NoQueryProvider(t *testing.T) {
	entity := &types.Entity{URN: "urn:li:dataset:test", Name: "Test"}
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return entity, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetEntity(context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetEntity (no query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetEntity_WithQueryProvider(t *testing.T) {
	entity := &types.Entity{URN: "urn:li:dataset:test", Name: "Test"}
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return entity, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(&mockQueryProvider{}))
	_, out, err := toolkit.handleGetEntity(context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetEntity (with query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetSchema_NoQueryProvider(t *testing.T) {
	schema := &types.SchemaMetadata{Name: "test_schema", Fields: []types.SchemaField{{FieldPath: "id", Type: "string"}}}
	mock := &mockClient{
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return schema, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetSchema(context.Background(), nil, GetSchemaInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetSchema (no query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetSchema_WithQueryProvider(t *testing.T) {
	schema := &types.SchemaMetadata{Name: "test_schema", Fields: []types.SchemaField{{FieldPath: "id", Type: "string"}}}
	mock := &mockClient{
		getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
			return schema, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(&mockQueryProvider{}))
	_, out, err := toolkit.handleGetSchema(context.Background(), nil, GetSchemaInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetSchema (with query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetLineage_NoQueryProvider(t *testing.T) {
	lineage := &types.LineageResult{
		Start:     "urn:li:dataset:test",
		Direction: "DOWNSTREAM",
		Nodes:     []types.LineageNode{{URN: "urn:li:dataset:up", Name: "upstream"}},
	}
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return lineage, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleGetLineage(context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetLineage (no query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_GetLineage_WithQueryProvider(t *testing.T) {
	lineage := &types.LineageResult{
		Start:     "urn:li:dataset:test",
		Direction: "DOWNSTREAM",
		Nodes:     []types.LineageNode{{URN: "urn:li:dataset:up", Name: "upstream"}},
	}
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return lineage, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(&mockQueryProvider{}))
	_, out, err := toolkit.handleGetLineage(context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleGetLineage (with query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_Search_NoQueryProvider(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Total:    1,
				Entities: []types.SearchEntity{{URN: "urn:li:dataset:test"}},
			}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	_, out, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleSearch (no query provider) structured output is nil, want non-nil")
	}
}

func TestStructuredOutput_Search_WithQueryProvider(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Total:    1,
				Entities: []types.SearchEntity{{URN: "urn:li:dataset:test"}},
			}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(&mockQueryProvider{}))
	_, out, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("handleSearch (with query provider) structured output is nil, want non-nil")
	}
}

// TestFormatJSONResult verifies formatJSONResult returns data as structured output.
func TestFormatJSONResult_ReturnsDataAsStructuredOutput(t *testing.T) {
	data := map[string]any{"key": "value"}
	_, out, err := formatJSONResult(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("formatJSONResult structured output is nil, want non-nil")
	}
	outMap, ok := out.(map[string]any)
	if !ok || outMap["key"] != "value" {
		t.Error("formatJSONResult should return input data as structured output")
	}
}

func TestFormatJSONResult_ErrorInput(t *testing.T) {
	// Channels can't be marshaled to JSON
	ch := make(chan int)
	result, out, err := formatJSONResult(ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("formatJSONResult with unmarshalable data should return error result")
	}
	if out != nil {
		t.Error("formatJSONResult with error should return nil structured output")
	}
}

// TestListDomainsOutput_WrapsDomainsCorrectly checks that the output wrapper
// preserves the domains slice exactly.
func TestListDomainsOutput_WrapsDomainsCorrectly(t *testing.T) {
	domains := []types.Domain{
		{URN: "urn:li:domain:a", Name: "A"},
		{URN: "urn:li:domain:b", Name: "B"},
	}
	output := ListDomainsOutput{Domains: domains}
	if len(output.Domains) != 2 {
		t.Errorf("ListDomainsOutput.Domains length = %d, want 2", len(output.Domains))
	}
}

// TestListTagsOutput_WrapsTagsCorrectly checks that the output wrapper preserves the tags slice.
func TestListTagsOutput_WrapsTagsCorrectly(t *testing.T) {
	tags := []types.Tag{{URN: "urn:li:tag:PII", Name: "PII"}}
	output := ListTagsOutput{Tags: tags}
	if len(output.Tags) != 1 {
		t.Errorf("ListTagsOutput.Tags length = %d, want 1", len(output.Tags))
	}
}

// TestListDataProductsOutput_WrapsProductsCorrectly checks the output wrapper.
func TestListDataProductsOutput_WrapsProductsCorrectly(t *testing.T) {
	products := []types.DataProduct{{URN: "urn:li:dataProduct:x", Name: "X"}}
	output := ListDataProductsOutput{DataProducts: products}
	if len(output.DataProducts) != 1 {
		t.Errorf("ListDataProductsOutput.DataProducts length = %d, want 1", len(output.DataProducts))
	}
}

// TestResponseShape_GetEntity_WithQueryProvider verifies that handleGetEntity with a query
// provider returns entity fields at the TOP LEVEL — not nested under an "entity" wrapper key.
// The Anthropic proxy rejects responses where the structured output shape diverges from OutputSchema.
func TestResponseShape_GetEntity_WithQueryProvider(t *testing.T) {
	entity := &types.Entity{URN: "urn:li:dataset:test", Name: "Test", Type: "DATASET"}
	mock := &mockClient{
		getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
			return entity, nil
		},
	}

	provider := &fullMockQueryProvider{
		resolveTableFn: func(_ context.Context, _ string) (*integration.TableIdentifier, error) {
			return &integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: "tbl"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))
	_, out, err := toolkit.handleGetEntity(context.Background(), nil, GetEntityInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("structured output is nil")
	}

	outMap, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("structured output type = %T, want map[string]any", out)
	}

	// Entity fields must be at top level — NOT nested under "entity"
	if _, hasEntityKey := outMap["entity"]; hasEntityKey {
		t.Error("response must not have an 'entity' wrapper key — entity fields must be at top level")
	}
	if _, hasURN := outMap["urn"]; !hasURN {
		t.Error("expected 'urn' at top level of response")
	}
	if _, hasQueryTable := outMap["query_table"]; !hasQueryTable {
		t.Error("expected 'query_table' at top level of response")
	}
}

// TestResponseShape_GetLineage_WithQueryProvider verifies that handleGetLineage with a query
// provider returns lineage fields at the TOP LEVEL — not nested under a "lineage" wrapper key.
func TestResponseShape_GetLineage_WithQueryProvider(t *testing.T) {
	lineage := &types.LineageResult{
		Start:     "urn:li:dataset:test",
		Direction: "DOWNSTREAM",
		Nodes:     []types.LineageNode{{URN: "urn:li:dataset:up", Name: "upstream"}},
	}
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return lineage, nil
		},
	}

	provider := &fullMockQueryProvider{
		getExecutionContextFn: func(_ context.Context, _ []string) (*integration.ExecutionContext, error) {
			return &integration.ExecutionContext{Source: "trino"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))
	_, out, err := toolkit.handleGetLineage(context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("structured output is nil")
	}

	outMap, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("structured output type = %T, want map[string]any", out)
	}

	// Lineage fields must be at top level — NOT nested under "lineage"
	if _, hasLineageKey := outMap["lineage"]; hasLineageKey {
		t.Error("response must not have a 'lineage' wrapper key — lineage fields must be at top level")
	}
	if _, hasStart := outMap["start"]; !hasStart {
		t.Error("expected 'start' at top level of response")
	}
	if _, hasNodes := outMap["nodes"]; !hasNodes {
		t.Error("expected 'nodes' at top level of response")
	}
	if _, hasExecCtx := outMap["execution_context"]; !hasExecCtx {
		t.Error("expected 'execution_context' at top level of response")
	}
}

// TestResponseShape_Search_WithQueryProvider verifies that formatSearchResult with a query
// provider returns search fields at the TOP LEVEL — not nested under a "result" wrapper key.
func TestResponseShape_Search_WithQueryProvider(t *testing.T) {
	mock := &mockClient{
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Total:    1,
				Entities: []types.SearchEntity{{URN: "urn:li:dataset:test", Name: "Test"}},
			}, nil
		},
	}

	provider := &fullMockQueryProvider{
		getTableAvailabilityFn: func(_ context.Context, _ string) (*integration.TableAvailability, error) {
			return &integration.TableAvailability{
				Available: true,
				Table:     &integration.TableIdentifier{Catalog: "cat", Schema: "sch", Table: "tbl"},
			}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig(), WithQueryProvider(provider))
	_, out, err := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("structured output is nil")
	}

	outMap, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("structured output type = %T, want map[string]any", out)
	}

	// Search fields must be at top level — NOT nested under "result"
	if _, hasResultKey := outMap["result"]; hasResultKey {
		t.Error("response must not have a 'result' wrapper key — search fields must be at top level")
	}
	if _, hasEntities := outMap["entities"]; !hasEntities {
		t.Error("expected 'entities' at top level of response")
	}
	if _, hasTotal := outMap["total"]; !hasTotal {
		t.Error("expected 'total' at top level of response")
	}
	if _, hasQueryCtx := outMap["query_context"]; !hasQueryCtx {
		t.Error("expected 'query_context' at top level of response")
	}
}
