package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleListDataProducts(t *testing.T) {
	tests := []struct {
		name         string
		mockProducts []types.DataProduct
		mockErr      error
		wantErr      bool
	}{
		{
			name: "successful list",
			mockProducts: []types.DataProduct{
				{URN: "urn:li:dataProduct:product1", Name: "Product 1", Description: "First product"},
				{URN: "urn:li:dataProduct:product2", Name: "Product 2", Description: "Second product"},
			},
			wantErr: false,
		},
		{
			name: "with domain",
			mockProducts: []types.DataProduct{
				{
					URN:    "urn:li:dataProduct:product1",
					Name:   "Product 1",
					Domain: &types.Domain{URN: "urn:li:domain:marketing", Name: "Marketing"},
				},
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
			result, _, _ := toolkit.handleListDataProducts(context.Background(), nil)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleListDataProducts() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleListDataProducts() should not return error result")
				}
			}
		})
	}
}

func TestHandleGetDataProduct(t *testing.T) {
	tests := []struct {
		name        string
		input       GetDataProductInput
		mockProduct *types.DataProduct
		mockErr     error
		wantErr     bool
	}{
		{
			name:  "successful get",
			input: GetDataProductInput{URN: "urn:li:dataProduct:test"},
			mockProduct: &types.DataProduct{
				URN:         "urn:li:dataProduct:test",
				Name:        "Test Product",
				Description: "A test data product",
			},
			wantErr: false,
		},
		{
			name:  "with full details",
			input: GetDataProductInput{URN: "urn:li:dataProduct:test"},
			mockProduct: &types.DataProduct{
				URN:         "urn:li:dataProduct:test",
				Name:        "Test Product",
				Description: "A test data product",
				Domain: &types.Domain{
					URN:  "urn:li:domain:marketing",
					Name: "Marketing",
				},
				Owners: []types.Owner{
					{URN: "urn:li:corpuser:john", Name: "John Doe", Type: "TECHNICAL_OWNER"},
				},
				Properties: map[string]string{
					"team": "data-engineering",
				},
			},
			wantErr: false,
		},
		{
			name:    "empty URN",
			input:   GetDataProductInput{URN: ""},
			wantErr: true,
		},
		{
			name:    "client error",
			input:   GetDataProductInput{URN: "urn:li:dataProduct:test"},
			mockErr: errors.New("not found"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getDataProductFunc: func(_ context.Context, _ string) (*types.DataProduct, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockProduct, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetDataProduct(context.Background(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Error("handleGetDataProduct() should return error result")
				}
			} else {
				if result.IsError {
					t.Error("handleGetDataProduct() should not return error result")
				}
			}
		})
	}
}
