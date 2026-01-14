package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleGetEntity(t *testing.T) {
	tests := []struct {
		name       string
		input      GetEntityInput
		mockEntity *types.Entity
		mockErr    error
		wantErr    bool
	}{
		{
			name:  "successful get",
			input: GetEntityInput{URN: "urn:li:dataset:test"},
			mockEntity: &types.Entity{
				URN:         "urn:li:dataset:test",
				Type:        "DATASET",
				Name:        "test",
				Description: "Test dataset",
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetEntityInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetEntityInput{URN: "urn:li:dataset:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getEntityFunc: func(_ context.Context, _ string) (*types.Entity, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockEntity, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetEntity(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetEntity() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetEntity() should not return error result")
				}
			}
		})
	}
}
