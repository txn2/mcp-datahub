package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleListDomains(t *testing.T) {
	tests := []struct {
		name        string
		mockDomains []types.Domain
		mockErr     error
		wantErr     bool
	}{
		{
			name: "successful list",
			mockDomains: []types.Domain{
				{URN: "urn:li:domain:marketing", Name: "Marketing", Description: "Marketing domain", EntityCount: 10},
				{URN: "urn:li:domain:sales", Name: "Sales", Description: "Sales domain", EntityCount: 20},
			},
			wantErr: false,
		},
		{
			name:        "empty list",
			mockDomains: []types.Domain{},
			wantErr:     false,
		},
		{
			name:    "client error",
			mockErr: errors.New("api error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				listDomainsFunc: func(_ context.Context) ([]types.Domain, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockDomains, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleListDomains(context.Background(), nil, ListDomainsInput{})

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleListDomains() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleListDomains() should not return error result")
				}
			}
		})
	}
}
