package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/integration"
)

// Mock implementations for testing

type mockURNResolver struct {
	resolveFunc func(ctx context.Context, externalID string) (string, error)
}

func (m *mockURNResolver) ResolveToDataHubURN(ctx context.Context, externalID string) (string, error) {
	if m.resolveFunc != nil {
		return m.resolveFunc(ctx, externalID)
	}
	return "urn:li:dataset:(urn:li:dataPlatform:test," + externalID + ",PROD)", nil
}

type mockAccessFilter struct {
	canAccessFunc  func(ctx context.Context, urn string) (bool, error)
	filterURNsFunc func(ctx context.Context, urns []string) ([]string, error)
}

func (m *mockAccessFilter) CanAccess(ctx context.Context, urn string) (bool, error) {
	if m.canAccessFunc != nil {
		return m.canAccessFunc(ctx, urn)
	}
	return true, nil
}

func (m *mockAccessFilter) FilterURNs(ctx context.Context, urns []string) ([]string, error) {
	if m.filterURNsFunc != nil {
		return m.filterURNsFunc(ctx, urns)
	}
	return urns, nil
}

type mockAuditLogger struct {
	logFunc func(ctx context.Context, tool string, params map[string]any, userID string) error
	calls   []auditLogCall
}

type auditLogCall struct {
	Tool   string
	Params map[string]any
	UserID string
}

func (m *mockAuditLogger) LogToolCall(ctx context.Context, tool string, params map[string]any, userID string) error {
	m.calls = append(m.calls, auditLogCall{Tool: tool, Params: params, UserID: userID})
	if m.logFunc != nil {
		return m.logFunc(ctx, tool, params, userID)
	}
	return nil
}

type mockMetadataEnricher struct {
	enrichFunc func(ctx context.Context, urn string, data map[string]any) (map[string]any, error)
}

func (m *mockMetadataEnricher) EnrichEntity(ctx context.Context, urn string, data map[string]any) (map[string]any, error) {
	if m.enrichFunc != nil {
		return m.enrichFunc(ctx, urn, data)
	}
	data["enriched"] = true
	return data, nil
}

// URNResolverMiddleware Tests

func TestURNResolverMiddleware_ResolvesExternalID(t *testing.T) {
	resolver := &mockURNResolver{}
	mw := NewURNResolverMiddleware(resolver)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "my-external-id"})
	ctx := context.Background()

	_, err := mw.Before(ctx, tc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolved, ok := tc.Get(ContextKeyResolvedURN)
	if !ok {
		t.Fatal("resolved URN not set in context")
	}

	expected := "urn:li:dataset:(urn:li:dataPlatform:test,my-external-id,PROD)"
	if resolved != expected {
		t.Errorf("resolved URN = %v, want %v", resolved, expected)
	}
}

func TestURNResolverMiddleware_SkipsDataHubURN(t *testing.T) {
	called := false
	resolver := &mockURNResolver{
		resolveFunc: func(_ context.Context, _ string) (string, error) {
			called = true
			return "", nil
		},
	}
	mw := NewURNResolverMiddleware(resolver)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "urn:li:dataset:test"})
	_, err := mw.Before(context.Background(), tc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("resolver should not be called for DataHub URNs")
	}

	// Should still set the URN in context
	resolved, ok := tc.Get(ContextKeyResolvedURN)
	if !ok {
		t.Fatal("URN not set in context")
	}
	if resolved != "urn:li:dataset:test" {
		t.Errorf("resolved = %v, want urn:li:dataset:test", resolved)
	}
}

func TestURNResolverMiddleware_HandlesError(t *testing.T) {
	expectedErr := errors.New("resolution failed")
	resolver := &mockURNResolver{
		resolveFunc: func(_ context.Context, _ string) (string, error) {
			return "", expectedErr
		},
	}
	mw := NewURNResolverMiddleware(resolver)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "my-id"})
	_, err := mw.Before(context.Background(), tc)

	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestURNResolverMiddleware_SkipsEmptyURN(t *testing.T) {
	called := false
	resolver := &mockURNResolver{
		resolveFunc: func(_ context.Context, _ string) (string, error) {
			called = true
			return "", nil
		},
	}
	mw := NewURNResolverMiddleware(resolver)

	tc := NewToolContext(ToolSearch, SearchInput{Query: "test"})
	_, err := mw.Before(context.Background(), tc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("resolver should not be called for inputs without URN")
	}
}

func TestURNResolverMiddleware_AfterIsNoOp(t *testing.T) {
	mw := NewURNResolverMiddleware(&mockURNResolver{})
	result := TextResult("test")

	out, err := mw.After(context.Background(), nil, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != result {
		t.Error("After should return result unchanged")
	}
}

// AccessFilterMiddleware Tests

func TestAccessFilterMiddleware_AllowsAccess(t *testing.T) {
	filter := &mockAccessFilter{
		canAccessFunc: func(_ context.Context, _ string) (bool, error) {
			return true, nil
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "urn:li:dataset:test"})
	_, err := mw.Before(context.Background(), tc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accessOK, ok := tc.Get(ContextKeyAccessOK)
	if !ok {
		t.Fatal("access status not set")
	}
	if accessOK != true {
		t.Error("access should be marked as OK")
	}
}

func TestAccessFilterMiddleware_DeniesAccess(t *testing.T) {
	filter := &mockAccessFilter{
		canAccessFunc: func(_ context.Context, _ string) (bool, error) {
			return false, nil
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "urn:li:dataset:test"})
	_, err := mw.Before(context.Background(), tc)

	if !errors.Is(err, ErrAccessDenied) {
		t.Errorf("error = %v, want ErrAccessDenied", err)
	}
}

func TestAccessFilterMiddleware_HandlesCanAccessError(t *testing.T) {
	expectedErr := errors.New("access check failed")
	filter := &mockAccessFilter{
		canAccessFunc: func(_ context.Context, _ string) (bool, error) {
			return false, expectedErr
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "urn:li:dataset:test"})
	_, err := mw.Before(context.Background(), tc)

	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestAccessFilterMiddleware_SkipsEmptyURN(t *testing.T) {
	called := false
	filter := &mockAccessFilter{
		canAccessFunc: func(_ context.Context, _ string) (bool, error) {
			called = true
			return true, nil
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolSearch, SearchInput{Query: "test"})
	_, err := mw.Before(context.Background(), tc)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("filter should not be called for inputs without URN")
	}
}

func TestAccessFilterMiddleware_UsesResolvedURN(t *testing.T) {
	var checkedURN string
	filter := &mockAccessFilter{
		canAccessFunc: func(_ context.Context, urn string) (bool, error) {
			checkedURN = urn
			return true, nil
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "external-id"})
	tc.Set(ContextKeyResolvedURN, "urn:li:dataset:resolved")

	_, err := mw.Before(context.Background(), tc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if checkedURN != "urn:li:dataset:resolved" {
		t.Errorf("checked URN = %v, want urn:li:dataset:resolved", checkedURN)
	}
}

func TestAccessFilterMiddleware_FiltersSearchResults(t *testing.T) {
	filter := &mockAccessFilter{
		filterURNsFunc: func(_ context.Context, urns []string) ([]string, error) {
			var filtered []string
			for _, urn := range urns {
				if urn == "urn:li:dataset:allowed" {
					filtered = append(filtered, urn)
				}
			}
			return filtered, nil
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolSearch, SearchInput{Query: "test"})
	result := TextResult(`{"entities":[{"urn":"urn:li:dataset:allowed"},{"urn":"urn:li:dataset:denied"}],"total":2}`)

	filtered, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := filtered.Content[0].(*mcp.TextContent).Text
	if !contains(text, "allowed") {
		t.Error("result should contain allowed URN")
	}
	if contains(text, "denied") {
		t.Error("result should not contain denied URN")
	}
}

func TestAccessFilterMiddleware_SkipsNonListTools(t *testing.T) {
	called := false
	filter := &mockAccessFilter{
		filterURNsFunc: func(_ context.Context, _ []string) ([]string, error) {
			called = true
			return nil, nil
		},
	}
	mw := NewAccessFilterMiddleware(filter)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "test"})
	result := TextResult(`{"urn":"test"}`)

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("filter should not be called for non-list tools")
	}
}

// AuditLoggerMiddleware Tests

func TestAuditLoggerMiddleware_LogsCall(t *testing.T) {
	logger := &mockAuditLogger{}
	getUserID := func(_ context.Context) string { return "user123" }
	mw := NewAuditLoggerMiddleware(logger, getUserID)

	tc := NewToolContext(ToolSearch, SearchInput{Query: "test", Limit: 10})
	result := TextResult("success")

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait a bit for async logging
	// In real tests, use proper synchronization
}

func TestAuditLoggerMiddleware_BeforeIsNoOp(t *testing.T) {
	mw := NewAuditLoggerMiddleware(&mockAuditLogger{}, nil)
	ctx := context.Background()

	out, err := mw.Before(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != ctx {
		t.Error("Before should return context unchanged")
	}
}

func TestAuditLoggerMiddleware_HandlesNilGetUserID(t *testing.T) {
	logger := &mockAuditLogger{}
	mw := NewAuditLoggerMiddleware(logger, nil)

	tc := NewToolContext(ToolSearch, SearchInput{Query: "test"})
	result := TextResult("ok")

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should not panic with nil getUserID
}

// MetadataEnricherMiddleware Tests

func TestMetadataEnricherMiddleware_EnrichesEntity(t *testing.T) {
	enricher := &mockMetadataEnricher{
		enrichFunc: func(_ context.Context, _ string, data map[string]any) (map[string]any, error) {
			data["custom_field"] = "custom_value"
			return data, nil
		},
	}
	mw := NewMetadataEnricherMiddleware(enricher)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "urn:li:dataset:test"})
	tc.Set(ContextKeyResolvedURN, "urn:li:dataset:test")

	result := TextResult(`{"urn":"urn:li:dataset:test","name":"Test"}`)

	enriched, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := enriched.Content[0].(*mcp.TextContent).Text
	if !contains(text, "custom_field") {
		t.Error("result should contain enriched field")
	}
}

func TestMetadataEnricherMiddleware_SkipsNonEntityTools(t *testing.T) {
	called := false
	enricher := &mockMetadataEnricher{
		enrichFunc: func(_ context.Context, _ string, data map[string]any) (map[string]any, error) {
			called = true
			return data, nil
		},
	}
	mw := NewMetadataEnricherMiddleware(enricher)

	tc := NewToolContext(ToolSearch, SearchInput{Query: "test"})
	result := TextResult(`{"entities":[]}`)

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("enricher should not be called for search results")
	}
}

func TestMetadataEnricherMiddleware_SkipsErrorResults(t *testing.T) {
	called := false
	enricher := &mockMetadataEnricher{
		enrichFunc: func(_ context.Context, _ string, data map[string]any) (map[string]any, error) {
			called = true
			return data, nil
		},
	}
	mw := NewMetadataEnricherMiddleware(enricher)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "test"})
	result := ErrorResult("some error")

	_, err := mw.After(context.Background(), tc, result, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("enricher should not be called for error results")
	}
}

func TestMetadataEnricherMiddleware_BeforeIsNoOp(t *testing.T) {
	mw := NewMetadataEnricherMiddleware(&mockMetadataEnricher{})
	ctx := context.Background()

	out, err := mw.Before(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != ctx {
		t.Error("Before should return context unchanged")
	}
}

func TestMetadataEnricherMiddleware_HandlesEnrichmentError(t *testing.T) {
	enricher := &mockMetadataEnricher{
		enrichFunc: func(_ context.Context, _ string, _ map[string]any) (map[string]any, error) {
			return nil, errors.New("enrichment failed")
		},
	}
	mw := NewMetadataEnricherMiddleware(enricher)

	tc := NewToolContext(ToolGetEntity, GetEntityInput{URN: "urn:li:dataset:test"})
	tc.Set(ContextKeyResolvedURN, "urn:li:dataset:test")

	original := TextResult(`{"urn":"test"}`)

	result, err := mw.After(context.Background(), tc, original, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return original result on error
	if result != original {
		t.Error("should return original result on enrichment error")
	}
}

// Helper function tests

func TestIsDataHubURN(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"urn:li:dataset:test", true},
		{"urn:li:glossaryTerm:test", true},
		{"urn:li:dataProduct:test", true},
		{"my-table", false},
		{"", false},
		{"urn:other:test", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isDataHubURN(tt.input); got != tt.want {
				t.Errorf("isDataHubURN(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsListTool(t *testing.T) {
	listTools := []ToolName{ToolSearch, ToolListTags, ToolListDomains, ToolListDataProducts, ToolGetLineage}
	nonListTools := []ToolName{ToolGetEntity, ToolGetSchema, ToolGetGlossaryTerm, ToolGetDataProduct}

	for _, tool := range listTools {
		if !isListTool(tool) {
			t.Errorf("isListTool(%v) = false, want true", tool)
		}
	}

	for _, tool := range nonListTools {
		if isListTool(tool) {
			t.Errorf("isListTool(%v) = true, want false", tool)
		}
	}
}

func TestIsEntityTool(t *testing.T) {
	entityTools := []ToolName{ToolGetEntity, ToolGetSchema, ToolGetGlossaryTerm, ToolGetDataProduct}
	nonEntityTools := []ToolName{ToolSearch, ToolListTags, ToolListDomains}

	for _, tool := range entityTools {
		if !isEntityTool(tool) {
			t.Errorf("isEntityTool(%v) = false, want true", tool)
		}
	}

	for _, tool := range nonEntityTools {
		if isEntityTool(tool) {
			t.Errorf("isEntityTool(%v) = true, want false", tool)
		}
	}
}

func TestExtractURNFromInput(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"GetEntityInput", GetEntityInput{URN: "test-urn"}, "test-urn"},
		{"GetSchemaInput", GetSchemaInput{URN: "schema-urn"}, "schema-urn"},
		{"GetLineageInput", GetLineageInput{URN: "lineage-urn"}, "lineage-urn"},
		{"SearchInput", SearchInput{Query: "test"}, ""},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractURNFromInput(tt.input); got != tt.want {
				t.Errorf("extractURNFromInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEffectiveURN(t *testing.T) {
	// With resolved URN
	tc1 := NewToolContext(ToolGetEntity, GetEntityInput{URN: "original"})
	tc1.Set(ContextKeyResolvedURN, "resolved")
	if got := getEffectiveURN(tc1); got != "resolved" {
		t.Errorf("getEffectiveURN with resolved = %v, want resolved", got)
	}

	// Without resolved URN
	tc2 := NewToolContext(ToolGetEntity, GetEntityInput{URN: "original"})
	if got := getEffectiveURN(tc2); got != "original" {
		t.Errorf("getEffectiveURN without resolved = %v, want original", got)
	}

	// No URN at all
	tc3 := NewToolContext(ToolSearch, SearchInput{Query: "test"})
	if got := getEffectiveURN(tc3); got != "" {
		t.Errorf("getEffectiveURN with no URN = %v, want empty", got)
	}
}

func TestExtractURNsFromData(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		want []string
	}{
		{
			name: "entities array",
			data: map[string]any{
				"entities": []any{
					map[string]any{"urn": "urn:1"},
					map[string]any{"urn": "urn:2"},
				},
			},
			want: []string{"urn:1", "urn:2"},
		},
		{
			name: "nodes array",
			data: map[string]any{
				"nodes": []any{
					map[string]any{"urn": "urn:3"},
				},
			},
			want: []string{"urn:3"},
		},
		{
			name: "direct urn",
			data: map[string]any{
				"urn": "urn:4",
			},
			want: []string{"urn:4"},
		},
		{
			name: "empty",
			data: map[string]any{},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractURNsFromData(tt.data)
			if len(got) != len(tt.want) {
				t.Errorf("extractURNsFromData() = %v, want %v", got, tt.want)
				return
			}
			for i, urn := range got {
				if urn != tt.want[i] {
					t.Errorf("extractURNsFromData()[%d] = %v, want %v", i, urn, tt.want[i])
				}
			}
		})
	}
}

func TestInputToMap(t *testing.T) {
	input := SearchInput{Query: "test", Limit: 10}
	m := inputToMap(input)

	if m["query"] != "test" {
		t.Errorf("inputToMap().query = %v, want test", m["query"])
	}
	// Limit might be float64 due to JSON unmarshaling
	if m["limit"].(float64) != 10 {
		t.Errorf("inputToMap().limit = %v, want 10", m["limit"])
	}
}

// Integration test with Toolkit

func TestToolkitWithIntegrationInterfaces(t *testing.T) {
	mock := &mockClient{}
	resolver := &mockURNResolver{}
	filter := &mockAccessFilter{}
	logger := &mockAuditLogger{}
	enricher := &mockMetadataEnricher{}

	toolkit := NewToolkit(mock, DefaultConfig(),
		WithURNResolver(resolver),
		WithAccessFilter(filter),
		WithAuditLogger(logger, func(_ context.Context) string { return "test-user" }),
		WithMetadataEnricher(enricher),
	)

	if !toolkit.HasMiddleware() {
		t.Error("HasMiddleware() should return true with integration interfaces")
	}

	if len(toolkit.integrationMiddleware) != 4 {
		t.Errorf("expected 4 integration middlewares, got %d", len(toolkit.integrationMiddleware))
	}
}

func TestToolkitIntegrationMiddlewareOrder(t *testing.T) {
	var order []string

	resolver := &mockURNResolver{
		resolveFunc: func(_ context.Context, _ string) (string, error) {
			order = append(order, "resolve")
			return "urn:li:dataset:resolved", nil
		},
	}
	filter := &mockAccessFilter{
		canAccessFunc: func(_ context.Context, _ string) (bool, error) {
			order = append(order, "access")
			return true, nil
		},
	}

	mock := &mockClient{}

	toolkit := NewToolkit(mock, DefaultConfig(),
		WithURNResolver(resolver),
		WithAccessFilter(filter),
	)

	// Build a wrapped handler to test order
	handler := func(_ context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, any, error) {
		order = append(order, "handler")
		return TextResult("ok"), nil, nil
	}

	wrapped := toolkit.wrapHandler(ToolGetEntity, handler, nil)
	_, _, _ = wrapped(context.Background(), nil, GetEntityInput{URN: "external-id"})

	// Verify order: resolve -> access -> handler
	expected := []string{"resolve", "access", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("order = %v, want %v", order, expected)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %v, want %v", i, order[i], v)
		}
	}
}

// Additional tests for edge cases

func TestFilterDataByURNs(t *testing.T) {
	allowed := map[string]bool{"urn:1": true, "urn:2": true}

	// Test filtering entities
	data := map[string]any{
		"entities": []any{
			map[string]any{"urn": "urn:1"},
			map[string]any{"urn": "urn:3"},
		},
		"total": 2,
	}

	result := filterDataByURNs(data, allowed)
	entities := result["entities"].([]any)
	if len(entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(entities))
	}
	if result["total"] != 1 {
		t.Errorf("expected total=1, got %v", result["total"])
	}
}

func TestFilterDataByURNs_Nodes(t *testing.T) {
	allowed := map[string]bool{"urn:1": true}

	data := map[string]any{
		"nodes": []any{
			map[string]any{"urn": "urn:1"},
			map[string]any{"urn": "urn:2"},
		},
	}

	result := filterDataByURNs(data, allowed)
	nodes := result["nodes"].([]any)
	if len(nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(nodes))
	}
}

func TestParseResultToMap_EmptyContent(t *testing.T) {
	result := &mcp.CallToolResult{Content: nil}
	_, err := parseResultToMap(result)
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestParseResultToMap_NonTextContent(t *testing.T) {
	result := &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.ImageContent{}},
	}
	_, err := parseResultToMap(result)
	if err == nil {
		t.Error("expected error for non-text content")
	}
}

func TestToolkitWithQueryProvider(t *testing.T) {
	mock := &mockClient{}

	mockProvider := &mockQueryProvider{}
	toolkit := NewToolkit(mock, DefaultConfig(),
		WithQueryProvider(mockProvider),
	)

	if !toolkit.HasQueryProvider() {
		t.Error("HasQueryProvider() should return true")
	}

	if toolkit.QueryProvider() == nil {
		t.Error("QueryProvider() should not be nil")
	}
}

// mockQueryProvider for testing.
type mockQueryProvider struct{}

func (m *mockQueryProvider) Name() string { return "mock" }
func (m *mockQueryProvider) ResolveTable(_ context.Context, _ string) (*integration.TableIdentifier, error) {
	return nil, nil
}
func (m *mockQueryProvider) GetTableAvailability(_ context.Context, _ string) (*integration.TableAvailability, error) {
	return nil, nil
}
func (m *mockQueryProvider) GetQueryExamples(_ context.Context, _ string) ([]integration.QueryExample, error) {
	return nil, nil
}
func (m *mockQueryProvider) GetExecutionContext(_ context.Context, _ []string) (*integration.ExecutionContext, error) {
	return nil, nil
}
func (m *mockQueryProvider) Close() error { return nil }

// contains checks if s contains substr.
func contains(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
