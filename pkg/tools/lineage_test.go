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

func TestHandleGetLineage_ColumnLevel(t *testing.T) {
	tests := []struct {
		name              string
		input             GetLineageInput
		mockColumnLineage *types.ColumnLineage
		mockErr           error
		wantErr           bool
	}{
		{
			name:  "column level with mappings",
			input: GetLineageInput{URN: "urn:li:dataset:test", Level: "column"},
			mockColumnLineage: &types.ColumnLineage{
				DatasetURN: "urn:li:dataset:test",
				Mappings: []types.ColumnLineageMapping{
					{
						DownstreamColumn: "user_id",
						UpstreamDataset:  "urn:li:dataset:source",
						UpstreamColumn:   "id",
						Transform:        "IDENTITY",
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "column level empty mappings",
			input: GetLineageInput{URN: "urn:li:dataset:test", Level: "column"},
			mockColumnLineage: &types.ColumnLineage{
				DatasetURN: "urn:li:dataset:test",
				Mappings:   []types.ColumnLineageMapping{},
			},
			wantErr: false,
		},
		{
			name:    "column level client error",
			input:   GetLineageInput{URN: "urn:li:dataset:test", Level: "column"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "column level empty URN",
			input:   GetLineageInput{URN: "", Level: "column"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getColumnLineageFunc: func(_ context.Context, _ string) (*types.ColumnLineage, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockColumnLineage, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetLineage(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetLineage(level=column) should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetLineage(level=column) should not return error result")
				}
			}
		})
	}
}

func TestHandleGetLineage_DefaultLevel(t *testing.T) {
	// When level is empty or "dataset", it should use dataset-level lineage
	mock := &mockClient{
		getLineageFunc: func(_ context.Context, _ string, _ ...client.LineageOption) (*types.LineageResult, error) {
			return &types.LineageResult{Start: "urn:li:dataset:test"}, nil
		},
	}

	toolkit := NewToolkit(mock, DefaultConfig())

	// Empty level (default)
	result, _, _ := toolkit.handleGetLineage(context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test"})
	if result.IsError {
		t.Error("empty level should use dataset-level lineage")
	}

	// Explicit "dataset" level
	result, _, _ = toolkit.handleGetLineage(context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test", Level: "dataset"})
	if result.IsError {
		t.Error("dataset level should use dataset-level lineage")
	}
}

func TestHandleGetLineage_InvalidLevel(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, _ := toolkit.handleGetLineage(context.Background(), nil, GetLineageInput{URN: "urn:li:dataset:test", Level: "invalid"})
	if !result.IsError {
		t.Error("invalid level should return error result")
	}
}
