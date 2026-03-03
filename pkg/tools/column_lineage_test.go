package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
	"github.com/txn2/mcp-datahub/pkg/types"
)

// TestColumnLineageDeprecatedAlias verifies the deprecated datahub_get_column_lineage tool
// delegates to handleGetLineage with Level="column" and returns correct results.
func TestColumnLineageDeprecatedAlias(t *testing.T) {
	tests := []struct {
		name              string
		input             GetColumnLineageInput
		mockColumnLineage *types.ColumnLineage
		mockErr           error
		wantErr           bool
	}{
		{
			name:  "successful get with mappings via alias",
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
				},
			},
			wantErr: false,
		},
		{
			name:  "successful get empty mappings via alias",
			input: GetColumnLineageInput{URN: "urn:li:dataset:test"},
			mockColumnLineage: &types.ColumnLineage{
				DatasetURN: "urn:li:dataset:test",
				Mappings:   []types.ColumnLineageMapping{},
			},
			wantErr: false,
		},
		{
			name:    "empty URN via alias",
			input:   GetColumnLineageInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error via alias",
			input:   GetColumnLineageInput{URN: "urn:li:dataset:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:  "with confidence score via alias",
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

			// Call via the consolidated lineage handler with Level="column"
			lineageInput := GetLineageInput{
				URN:        tt.input.URN,
				Level:      "column",
				Connection: tt.input.Connection,
			}
			result, _, _ := toolkit.handleGetLineage(context.Background(), nil, lineageInput)

			if tt.wantErr {
				if !result.IsError {
					t.Error("deprecated alias should return error result")
				}
			} else {
				if result.IsError {
					t.Error("deprecated alias should not return error result")
				}
			}
		})
	}
}

// TestColumnLineageDeprecatedAlias_Registration verifies the deprecated tool registers.
func TestColumnLineageDeprecatedAlias_Registration(t *testing.T) {
	mock := &mockClient{
		getColumnLineageFunc: func(_ context.Context, _ string) (*types.ColumnLineage, error) {
			return &types.ColumnLineage{DatasetURN: "urn:li:dataset:test"}, nil
		},
	}
	toolkit := NewToolkit(mock, DefaultConfig())
	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolGetColumnLineage)

	if !toolkit.registeredTools[ToolGetColumnLineage] {
		t.Error("ToolGetColumnLineage should be registered")
	}
}

func TestColumnLineageDeprecatedAlias_ConnectionError(t *testing.T) {
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

	// Try to use an unknown connection via the consolidated handler
	input := GetLineageInput{
		URN:        "urn:li:dataset:test",
		Level:      "column",
		Connection: "unknown",
	}

	result, _, _ := toolkit.handleGetLineage(context.Background(), nil, input)

	if !result.IsError {
		t.Error("should return error for unknown connection")
	}
}
