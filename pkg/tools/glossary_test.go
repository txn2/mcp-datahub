package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleGetGlossaryTerm(t *testing.T) {
	tests := []struct {
		name     string
		input    GetGlossaryTermInput
		mockTerm *types.GlossaryTerm
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful get",
			input: GetGlossaryTermInput{URN: "urn:li:glossaryTerm:business.revenue"},
			mockTerm: &types.GlossaryTerm{
				URN:         "urn:li:glossaryTerm:business.revenue",
				Name:        "Revenue",
				Description: "Total revenue",
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetGlossaryTermInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetGlossaryTermInput{URN: "urn:li:glossaryTerm:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getGlossaryTermFunc: func(_ context.Context, _ string) (*types.GlossaryTerm, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockTerm, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetGlossaryTerm(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetGlossaryTerm() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetGlossaryTerm() should not return error result")
				}
			}
		})
	}
}
