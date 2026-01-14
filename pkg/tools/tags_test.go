package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleListTags(t *testing.T) {
	tests := []struct {
		name     string
		input    ListTagsInput
		mockTags []types.Tag
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful list",
			input: ListTagsInput{},
			mockTags: []types.Tag{
				{URN: "urn:li:tag:PII", Name: "PII", Description: "Personal info"},
				{URN: "urn:li:tag:Sensitive", Name: "Sensitive", Description: "Sensitive data"},
			},
			wantErr: false,
		},
		{
			name:  "with filter",
			input: ListTagsInput{Filter: "PII"},
			mockTags: []types.Tag{
				{URN: "urn:li:tag:PII", Name: "PII"},
			},
			wantErr: false,
		},
		{
			name:     "empty list",
			input:    ListTagsInput{},
			mockTags: []types.Tag{},
			wantErr:  false,
		},
		{
			name:    "client error",
			input:   ListTagsInput{},
			mockErr: errors.New("api error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				listTagsFunc: func(_ context.Context, _ string) ([]types.Tag, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockTags, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleListTags(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleListTags() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleListTags() should not return error result")
				}
			}
		})
	}
}
