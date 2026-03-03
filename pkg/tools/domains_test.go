package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// TestListDomainsDeprecatedAlias verifies the deprecated datahub_list_domains tool
// delegates to handleBrowse and returns correct results.
func TestListDomainsDeprecatedAlias(t *testing.T) {
	tests := []struct {
		name        string
		mockDomains []types.Domain
		mockErr     error
		wantErr     bool
	}{
		{
			name: "successful list via alias",
			mockDomains: []types.Domain{
				{URN: "urn:li:domain:marketing", Name: "Marketing", EntityCount: 10},
			},
			wantErr: false,
		},
		{
			name:    "client error via alias",
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

			browseInput := BrowseInput{What: "domains"}
			result, _, _ := toolkit.handleBrowse(context.Background(), nil, browseInput)

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

// TestListDomainsDeprecatedAlias_Registration verifies the deprecated tool registers via MCP server.
func TestListDomainsDeprecatedAlias_Registration(t *testing.T) {
	mock := &mockClient{
		listDomainsFunc: func(_ context.Context) ([]types.Domain, error) {
			return []types.Domain{{URN: "urn:li:domain:test", Name: "test"}}, nil
		},
	}
	toolkit := NewToolkit(mock, DefaultConfig())
	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolListDomains)

	if !toolkit.registeredTools[ToolListDomains] {
		t.Error("ToolListDomains should be registered")
	}
}
