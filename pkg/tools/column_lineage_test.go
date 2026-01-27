package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleGetColumnLineage(t *testing.T) {
	tests := []struct {
		name              string
		input             GetColumnLineageInput
		mockColumnLineage *types.ColumnLineage
		mockErr           error
		wantErr           bool
	}{
		{
			name:  "successful get with mappings",
			input: GetColumnLineageInput{URN: "urn:li:dataset:test"},
			mockColumnLineage: &types.ColumnLineage{
				DatasetURN: "urn:li:dataset:test",
				Mappings: []types.ColumnLineageMapping{
					{
						DownstreamColumn: "user_id",
						UpstreamDataset:  "urn:li:dataset:source",
						UpstreamColumn:   "id",
						Transform:        "IDENTITY",
					},
					{
						DownstreamColumn: "full_name",
						UpstreamDataset:  "urn:li:dataset:source",
						UpstreamColumn:   "name",
						Transform:        "TRANSFORM",
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "successful get empty mappings",
			input: GetColumnLineageInput{URN: "urn:li:dataset:test"},
			mockColumnLineage: &types.ColumnLineage{
				DatasetURN: "urn:li:dataset:test",
				Mappings:   []types.ColumnLineageMapping{},
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetColumnLineageInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetColumnLineageInput{URN: "urn:li:dataset:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:  "successful get with confidence score and query",
			input: GetColumnLineageInput{URN: "urn:li:dataset:test"},
			mockColumnLineage: &types.ColumnLineage{
				DatasetURN: "urn:li:dataset:test",
				Mappings: []types.ColumnLineageMapping{
					{
						DownstreamColumn: "total",
						UpstreamDataset:  "urn:li:dataset:source",
						UpstreamColumn:   "amount",
						Transform:        "AGGREGATE",
						ConfidenceScore:  0.95,
						Query:            "SELECT SUM(amount) as total FROM source",
					},
				},
			},
			wantErr: false,
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
			result, _, _ := toolkit.handleGetColumnLineage(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetColumnLineage() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetColumnLineage() should not return error result")
				}
			}
		})
	}
}

func TestHandleGetColumnLineage_ConnectionError(t *testing.T) {
	cfg := multiserver.Config{
		Default: "prod",
		Primary: client.Config{
			URL:   "https://prod.datahub.example.com",
			Token: "prod-token",
		},
	}
	mgr := multiserver.NewManager(cfg)
	defer func() {
		_ = mgr.Close()
	}()

	toolkit := NewToolkitWithManager(mgr, DefaultConfig())

	// Try to use an unknown connection
	input := GetColumnLineageInput{
		URN:        "urn:li:dataset:test",
		Connection: "unknown",
	}

	result, _, _ := toolkit.handleGetColumnLineage(context.Background(), nil, input)

	if !result.IsError {
		t.Error("handleGetColumnLineage() should return error for unknown connection")
	}
}
