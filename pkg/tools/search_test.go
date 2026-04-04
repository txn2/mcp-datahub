package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleSearch(t *testing.T) {
	tests := []struct {
		name       string
		input      SearchInput
		mockResult *types.SearchResult
		mockErr    error
		wantErr    bool
	}{
		{
			name:  "successful search",
			input: SearchInput{Query: "test"},
			mockResult: &types.SearchResult{
				Total: 1,
				Entities: []types.SearchEntity{
					{URN: "urn:li:dataset:test", Name: "test"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty query",
			input:   SearchInput{Query: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   SearchInput{Query: "test"},
			mockErr: errors.New("client error"),
			wantErr: true,
		},
		{
			name: "with entity type",
			input: SearchInput{
				Query:      "dashboard",
				EntityType: "DASHBOARD",
			},
			mockResult: &types.SearchResult{
				Total: 1,
				Entities: []types.SearchEntity{
					{URN: "urn:li:dashboard:test", Name: "dashboard", Type: "DASHBOARD"},
				},
			},
			wantErr: false,
		},
		{
			name: "with pagination",
			input: SearchInput{
				Query:  "test",
				Limit:  20,
				Offset: 10,
			},
			mockResult: &types.SearchResult{
				Total:  100,
				Offset: 10,
				Limit:  20,
			},
			wantErr: false,
		},
		{
			name: "with types override",
			input: SearchInput{
				Query: "revenue",
				Types: []string{"DATASET", "DASHBOARD"},
			},
			mockResult: &types.SearchResult{
				Total: 2,
				Entities: []types.SearchEntity{
					{URN: "urn:li:dataset:rev", Name: "revenue", Type: "DATASET"},
					{URN: "urn:li:dashboard:rev", Name: "revenue-dash", Type: "DASHBOARD"},
				},
			},
			wantErr: false,
		},
		{
			name: "with filters",
			input: SearchInput{
				Query: "*",
				Filters: []SearchFilterInput{
					{Field: "fieldPaths", Values: []string{"email"}, Condition: "CONTAIN"},
					{Field: "platform", Value: "urn:li:dataPlatform:trino"},
				},
			},
			mockResult: &types.SearchResult{
				Total: 1,
				Entities: []types.SearchEntity{
					{URN: "urn:li:dataset:users", Name: "users"},
				},
			},
			wantErr: false,
		},
		{
			name: "filter with empty field",
			input: SearchInput{
				Query:   "*",
				Filters: []SearchFilterInput{{Field: "", Value: "trino"}},
			},
			wantErr: true,
		},
		{
			name: "filter with no values",
			input: SearchInput{
				Query:   "*",
				Filters: []SearchFilterInput{{Field: "platform"}},
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			input: SearchInput{
				Query: "test",
				Mode:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				searchAcrossEntitiesFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockResult, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleSearch(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleSearch() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleSearch() should not return error result")
				}
			}
		})
	}
}

func TestHandleSearch_SemanticMode(t *testing.T) {
	semanticCalled := false
	mock := &mockClient{
		semanticSearchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			semanticCalled = true
			return &types.SearchResult{Total: 1}, nil
		},
		searchAcrossEntitiesFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			t.Error("SearchAcrossEntities should not be called in semantic mode")
			return &types.SearchResult{}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	result, _, _ := toolkit.handleSearch(context.Background(), nil, SearchInput{
		Query: "test",
		Mode:  "semantic",
	})

	if result.IsError {
		t.Error("handleSearch() should not return error result")
	}
	if !semanticCalled {
		t.Error("SemanticSearch should have been called")
	}
}

func TestHandleSearch_KeywordUsesSearchAcrossEntities(t *testing.T) {
	acrossCalled := false
	mock := &mockClient{
		searchAcrossEntitiesFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			acrossCalled = true
			return &types.SearchResult{Total: 0}, nil
		},
		searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			t.Error("Search should not be called for keyword mode")
			return &types.SearchResult{}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	result, _, _ := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "test"})

	if result.IsError {
		t.Error("handleSearch() should not return error result")
	}
	if !acrossCalled {
		t.Error("SearchAcrossEntities should have been called")
	}
}

func TestHandleSearch_ZeroResults_EntitiesIsEmptyArray(t *testing.T) {
	// Mock returns initialized empty slice (as the fixed client now does).
	// The real nil→[] fix is tested at the client level in
	// TestSearchAcrossEntities_ZeroResults_EntitiesNotNil.
	mock := &mockClient{
		searchAcrossEntitiesFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
			return &types.SearchResult{
				Entities: []types.SearchEntity{},
				Total:    0,
			}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	result, _, _ := toolkit.handleSearch(context.Background(), nil, SearchInput{Query: "nonexistent"})

	if result.IsError {
		t.Fatal("handleSearch() should not return error result")
	}

	// Extract JSON text and verify entities is [] not null
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}

	if strings.Contains(tc.Text, `"entities": null`) {
		t.Error("entities should be [] not null in JSON output")
	}
	if !strings.Contains(tc.Text, `"entities": []`) {
		t.Error("expected empty entities array in JSON output")
	}
}

func TestSearchInputValidation(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	// Test with invalid input type (simulating type assertion failure)
	baseHandler := func(_ context.Context, _ *mcp.CallToolRequest, input any) (*mcp.CallToolResult, any, error) {
		searchInput, ok := input.(SearchInput)
		if !ok {
			return ErrorResult("internal error: invalid input type"), nil, nil
		}
		return toolkit.handleSearch(context.Background(), nil, searchInput)
	}

	// Pass wrong type
	result, _, _ := baseHandler(context.Background(), nil, "wrong type")
	if !result.IsError {
		t.Error("Should return error for invalid input type")
	}
}

func TestResolveSearchTypes(t *testing.T) {
	tests := []struct {
		name  string
		input SearchInput
		want  []string
	}{
		{
			name:  "default to DATASET",
			input: SearchInput{Query: "test"},
			want:  []string{"DATASET"},
		},
		{
			name:  "entity_type set",
			input: SearchInput{Query: "test", EntityType: "DASHBOARD"},
			want:  []string{"DASHBOARD"},
		},
		{
			name:  "types overrides entity_type",
			input: SearchInput{Query: "test", EntityType: "TAG", Types: []string{"DATASET", "DASHBOARD"}},
			want:  []string{"DATASET", "DASHBOARD"},
		},
		{
			name:  "types set without entity_type",
			input: SearchInput{Query: "test", Types: []string{"DATA_PRODUCT"}},
			want:  []string{"DATA_PRODUCT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSearchTypes(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("resolveSearchTypes() = %v, want %v", got, tt.want)
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("resolveSearchTypes()[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestConvertFilters(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []SearchFilterInput
		wantLen int
	}{
		{
			name:    "empty filters",
			inputs:  nil,
			wantLen: 0,
		},
		{
			name: "value promoted to values",
			inputs: []SearchFilterInput{
				{Field: "platform", Value: "urn:li:dataPlatform:trino"},
			},
			wantLen: 1,
		},
		{
			name: "values used directly",
			inputs: []SearchFilterInput{
				{Field: "fieldPaths", Values: []string{"email", "phone"}},
			},
			wantLen: 1,
		},
		{
			name: "with condition and negated",
			inputs: []SearchFilterInput{
				{Field: "tags", Values: []string{"urn:li:tag:deprecated"}, Condition: "EQUAL", Negated: true},
			},
			wantLen: 1,
		},
		{
			name: "both value and values merges all",
			inputs: []SearchFilterInput{
				{Field: "fieldPaths", Value: "email", Values: []string{"phone", "address"}},
			},
			wantLen: 1,
		},
		{
			name: "value already in values is not duplicated",
			inputs: []SearchFilterInput{
				{Field: "fieldPaths", Value: "email", Values: []string{"email", "phone"}},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertFilters(tt.inputs)
			if len(got) != tt.wantLen {
				t.Fatalf("convertFilters() len = %d, want %d", len(got), tt.wantLen)
			}

			// Verify value promotion
			if tt.name == "value promoted to values" {
				if len(got[0].Values) != 1 || got[0].Values[0] != "urn:li:dataPlatform:trino" {
					t.Errorf("value not promoted: %v", got[0].Values)
				}
			}

			// Verify condition/negated pass-through
			if tt.name == "with condition and negated" {
				if got[0].Condition != "EQUAL" {
					t.Errorf("condition = %q, want EQUAL", got[0].Condition)
				}
				if !got[0].Negated {
					t.Error("negated should be true")
				}
			}

			// Verify both value+values are merged
			if tt.name == "both value and values merges all" {
				want := []string{"email", "phone", "address"}
				if len(got[0].Values) != 3 {
					t.Fatalf("expected 3 values, got %d: %v", len(got[0].Values), got[0].Values)
				}
				for i, v := range got[0].Values {
					if v != want[i] {
						t.Errorf("values[%d] = %q, want %q", i, v, want[i])
					}
				}
			}

			// Verify duplicate is not added
			if tt.name == "value already in values is not duplicated" {
				want := []string{"email", "phone"}
				if len(got[0].Values) != 2 {
					t.Fatalf("expected 2 values, got %d: %v", len(got[0].Values), got[0].Values)
				}
				for i, v := range got[0].Values {
					if v != want[i] {
						t.Errorf("values[%d] = %q, want %q", i, v, want[i])
					}
				}
			}
		})
	}
}

func TestValidateFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters []SearchFilterInput
		wantErr bool
	}{
		{
			name:    "nil filters",
			filters: nil,
			wantErr: false,
		},
		{
			name:    "valid filter with value",
			filters: []SearchFilterInput{{Field: "platform", Value: "trino"}},
			wantErr: false,
		},
		{
			name:    "valid filter with values",
			filters: []SearchFilterInput{{Field: "fieldPaths", Values: []string{"email"}}},
			wantErr: false,
		},
		{
			name:    "empty field",
			filters: []SearchFilterInput{{Field: "", Value: "trino"}},
			wantErr: true,
		},
		{
			name:    "empty value and values",
			filters: []SearchFilterInput{{Field: "platform"}},
			wantErr: true,
		},
		{
			name: "second filter invalid",
			filters: []SearchFilterInput{
				{Field: "platform", Value: "trino"},
				{Field: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilters(tt.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFilters() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
