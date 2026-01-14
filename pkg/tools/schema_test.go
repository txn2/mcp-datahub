package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleGetSchema(t *testing.T) {
	tests := []struct {
		name       string
		input      GetSchemaInput
		mockSchema *types.SchemaMetadata
		mockErr    error
		wantErr    bool
	}{
		{
			name:  "successful get",
			input: GetSchemaInput{URN: "urn:li:dataset:test"},
			mockSchema: &types.SchemaMetadata{
				Name:    "schema",
				Version: 1,
				Fields: []types.SchemaField{
					{FieldPath: "id", Type: "NUMBER"},
					{FieldPath: "name", Type: "STRING"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetSchemaInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetSchemaInput{URN: "urn:li:dataset:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getSchemaFunc: func(_ context.Context, _ string) (*types.SchemaMetadata, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockSchema, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetSchema(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetSchema() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetSchema() should not return error result")
				}
			}
		})
	}
}
