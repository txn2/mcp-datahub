package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleGetLineage(t *testing.T) {
	tests := []struct {
		name        string
		input       GetLineageInput
		mockLineage *types.LineageResult
		mockErr     error
		wantErr     bool
	}{
		{
			name:  "successful get downstream",
			input: GetLineageInput{URN: "urn:li:dataset:test"},
			mockLineage: &types.LineageResult{
				Start:     "urn:li:dataset:test",
				Direction: "DOWNSTREAM",
				Depth:     1,
				Nodes: []types.LineageNode{
					{URN: "urn:li:dataset:downstream", Name: "downstream"},
				},
			},
			wantErr: false,
		},
		{
			name: "successful get upstream",
			input: GetLineageInput{
				URN:       "urn:li:dataset:test",
				Direction: "UPSTREAM",
			},
			mockLineage: &types.LineageResult{
				Start:     "urn:li:dataset:test",
				Direction: "UPSTREAM",
				Depth:     1,
				Nodes: []types.LineageNode{
					{URN: "urn:li:dataset:upstream", Name: "upstream"},
				},
			},
			wantErr: false,
		},
		{
			name: "with depth",
			input: GetLineageInput{
				URN:   "urn:li:dataset:test",
				Depth: 3,
			},
			mockLineage: &types.LineageResult{
				Start: "urn:li:dataset:test",
				Depth: 3,
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetLineageInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetLineageInput{URN: "urn:li:dataset:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockLineage, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetLineage(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetLineage() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetLineage() should not return error result")
				}
			}
		})
	}
}
