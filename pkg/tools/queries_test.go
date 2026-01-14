package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleGetQueries(t *testing.T) {
	tests := []struct {
		name        string
		input       GetQueriesInput
		mockQueries *types.QueryList
		mockErr     error
		wantErr     bool
	}{
		{
			name:  "successful get with queries",
			input: GetQueriesInput{URN: "urn:li:dataset:test"},
			mockQueries: &types.QueryList{
				Total: 2,
				Queries: []types.Query{
					{Statement: "SELECT * FROM table"},
					{Statement: "SELECT id FROM table WHERE active = true"},
				},
			},
			wantErr: false,
		},
		{
			name:  "successful get empty",
			input: GetQueriesInput{URN: "urn:li:dataset:test"},
			mockQueries: &types.QueryList{
				Total:   0,
				Queries: nil,
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetQueriesInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetQueriesInput{URN: "urn:li:dataset:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getQueriesFunc: func(_ context.Context, _ string) (*types.QueryList, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockQueries, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetQueries(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetQueries() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetQueries() should not return error result")
				}
			}
		})
	}
}
