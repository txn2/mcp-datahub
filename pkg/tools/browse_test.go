package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/client"
	"github.com/txn2/mcp-datahub/pkg/multiserver"
	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleBrowse_Tags(t *testing.T) {
	tests := []struct {
		name     string
		input    BrowseInput
		mockTags []types.Tag
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful list",
			input: BrowseInput{What: "tags"},
			mockTags: []types.Tag{
				{URN: "urn:li:tag:PII", Name: "PII", Description: "Personal info"},
				{URN: "urn:li:tag:Sensitive", Name: "Sensitive", Description: "Sensitive data"},
			},
			wantErr: false,
		},
		{
			name:  "with filter",
			input: BrowseInput{What: "tags", Filter: "PII"},
			mockTags: []types.Tag{
				{URN: "urn:li:tag:PII", Name: "PII"},
			},
			wantErr: false,
		},
		{
			name:     "empty list",
			input:    BrowseInput{What: "tags"},
			mockTags: []types.Tag{},
			wantErr:  false,
		},
		{
			name:    "client error",
			input:   BrowseInput{What: "tags"},
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
			result, _, _ := toolkit.handleBrowse(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleBrowse() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleBrowse() should not return error result")
				}
			}
		})
	}
}

func TestHandleBrowse_Domains(t *testing.T) {
	tests := []struct {
		name        string
		mockDomains []types.Domain
		mockErr     error
		wantErr     bool
	}{
		{
			name: "successful list",
			mockDomains: []types.Domain{
				{URN: "urn:li:domain:marketing", Name: "Marketing", EntityCount: 10},
				{URN: "urn:li:domain:sales", Name: "Sales", EntityCount: 20},
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
			result, _, _ := toolkit.handleBrowse(context.Background(), nil, BrowseInput{What: "domains"})

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleBrowse() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleBrowse() should not return error result")
				}
			}
		})
	}
}

func TestHandleBrowse_DataProducts(t *testing.T) {
	tests := []struct {
		name         string
		mockProducts []types.DataProduct
		mockErr      error
		wantErr      bool
	}{
		{
			name: "successful list",
			mockProducts: []types.DataProduct{
				{URN: "urn:li:dataProduct:product1", Name: "Product 1"},
				{URN: "urn:li:dataProduct:product2", Name: "Product 2"},
			},
			wantErr: false,
		},
		{
			name:         "empty list",
			mockProducts: []types.DataProduct{},
			wantErr:      false,
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
				listDataProductsFunc: func(_ context.Context) ([]types.DataProduct, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockProducts, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleBrowse(context.Background(), nil, BrowseInput{What: "data_products"})

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleBrowse() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleBrowse() should not return error result")
				}
			}
		})
	}
}

func TestHandleBrowse_InvalidWhat(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, _ := toolkit.handleBrowse(context.Background(), nil, BrowseInput{What: "invalid"})
	if !result.IsError {
		t.Error("handleBrowse() should return error for invalid what")
	}
}

func TestHandleBrowse_EmptyWhat(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig())

	result, _, _ := toolkit.handleBrowse(context.Background(), nil, BrowseInput{})
	if !result.IsError {
		t.Error("handleBrowse() should return error for empty what")
	}
}

func TestHandleBrowse_ConnectionError(t *testing.T) {
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

	input := BrowseInput{
		What:       "tags",
		Connection: "unknown",
	}

	result, _, _ := toolkit.handleBrowse(context.Background(), nil, input)

	if !result.IsError {
		t.Error("handleBrowse() should return error for unknown connection")
	}
}
