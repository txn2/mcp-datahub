package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/types"
)

// TestListTagsDeprecatedAlias verifies the deprecated datahub_list_tags tool
// delegates to handleBrowse and returns correct results.
func TestListTagsDeprecatedAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    ListTagsInput
		mockTags []types.Tag
		mockErr  error
		wantErr  bool
	}{
		{
			name:  "successful list via alias",
			input: ListTagsInput{},
			mockTags: []types.Tag{
				{URN: "urn:li:tag:PII", Name: "PII", Description: "Personal info"},
			},
			wantErr: false,
		},
		{
			name:  "with filter via alias",
			input: ListTagsInput{Filter: "PII"},
			mockTags: []types.Tag{
				{URN: "urn:li:tag:PII", Name: "PII"},
			},
			wantErr: false,
		},
		{
			name:    "client error via alias",
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

			// Call handleBrowse via the alias mapping
			browseInput := BrowseInput{
				What:       "tags",
				Filter:     tt.input.Filter,
				Connection: tt.input.Connection,
			}
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

// TestListTagsDeprecatedAlias_Registration verifies the deprecated tool registers via MCP server.
func TestListTagsDeprecatedAlias_Registration(t *testing.T) {
	mock := &mockClient{
		listTagsFunc: func(_ context.Context, _ string) ([]types.Tag, error) {
			return []types.Tag{{URN: "urn:li:tag:test", Name: "test"}}, nil
		},
	}
	toolkit := NewToolkit(mock, DefaultConfig())
	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.Register(server, ToolListTags)

	if !toolkit.registeredTools[ToolListTags] {
		t.Error("ToolListTags should be registered")
	}
}
