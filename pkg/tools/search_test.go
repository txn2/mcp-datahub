package tools

import (
	"context"
	"errors"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				searchFunc: func(_ context.Context, _ string, _ ...client.SearchOption) (*types.SearchResult, error) {
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
