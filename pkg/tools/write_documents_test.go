package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleCreateDocument(t *testing.T) {
	tests := []struct {
		name       string
		input      CreateDocumentInput
		mockFunc   func(ctx context.Context, input types.CreateDocumentInput) (string, error)
		wantErr    bool
		wantErrMsg string
		wantURN    string
	}{
		{
			name:  "success",
			input: CreateDocumentInput{Title: "My Doc", Content: "Hello"},
			mockFunc: func(_ context.Context, input types.CreateDocumentInput) (string, error) {
				if input.Title != "My Doc" {
					return "", errors.New("unexpected title")
				}
				return "urn:li:document:new-1", nil
			},
			wantURN: "urn:li:document:new-1",
		},
		{
			name:       "missing title",
			input:      CreateDocumentInput{Content: "Hello"},
			wantErr:    true,
			wantErrMsg: "title parameter is required",
		},
		{
			name:       "missing content",
			input:      CreateDocumentInput{Title: "My Doc"},
			wantErr:    true,
			wantErrMsg: "content parameter is required",
		},
		{
			name:  "client error",
			input: CreateDocumentInput{Title: "My Doc", Content: "Hello"},
			mockFunc: func(_ context.Context, _ types.CreateDocumentInput) (string, error) {
				return "", errors.New("permission denied")
			},
			wantErr:    true,
			wantErrMsg: "CreateDocument failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{}
			if tt.mockFunc != nil {
				mock.createDocumentFunc = tt.mockFunc
			}
			toolkit := NewToolkit(mock, Config{WriteEnabled: true})

			result, _, _ := toolkit.handleCreateDocument(t.Context(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Fatal("expected error result")
				}
				if tt.wantErrMsg != "" {
					assertResultContains(t, result, tt.wantErrMsg)
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result")
			}
		})
	}
}

func TestHandleUpdateDocument(t *testing.T) {
	tests := []struct {
		name           string
		input          UpdateDocumentInput
		wantErr        bool
		wantErrMsg     string
		contentsCalled bool
		statusCalled   bool
		assetsCalled   bool
	}{
		{
			name:           "update title and content",
			input:          UpdateDocumentInput{URN: "urn:li:document:doc-1", Title: "New", Content: "Body"},
			contentsCalled: true,
		},
		{
			name:         "update status only",
			input:        UpdateDocumentInput{URN: "urn:li:document:doc-1", State: "UNPUBLISHED"},
			statusCalled: true,
		},
		{
			name:         "update related assets",
			input:        UpdateDocumentInput{URN: "urn:li:document:doc-1", RelatedAssets: []string{"urn:li:dataset:ds1"}},
			assetsCalled: true,
		},
		{
			name:       "missing urn",
			input:      UpdateDocumentInput{Title: "New"},
			wantErr:    true,
			wantErrMsg: "urn parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var contentsCalled, statusCalled, assetsCalled bool

			mock := &mockClient{}
			mock.updateDocContFunc = func(_ context.Context, _, _, _ string) error {
				contentsCalled = true
				return nil
			}
			mock.updateDocStatusFunc = func(_ context.Context, _, _ string) error {
				statusCalled = true
				return nil
			}
			mock.updateDocRelFunc = func(_ context.Context, _ string, _ []string) error {
				assetsCalled = true
				return nil
			}

			toolkit := NewToolkit(mock, Config{WriteEnabled: true})
			result, _, _ := toolkit.handleUpdateDocument(t.Context(), nil, tt.input)

			if tt.wantErr {
				if !result.IsError {
					t.Fatal("expected error result")
				}
				if tt.wantErrMsg != "" {
					assertResultContains(t, result, tt.wantErrMsg)
				}
				return
			}

			if result.IsError {
				t.Fatalf("unexpected error result")
			}

			if contentsCalled != tt.contentsCalled {
				t.Errorf("contentsCalled = %v, want %v", contentsCalled, tt.contentsCalled)
			}
			if statusCalled != tt.statusCalled {
				t.Errorf("statusCalled = %v, want %v", statusCalled, tt.statusCalled)
			}
			if assetsCalled != tt.assetsCalled {
				t.Errorf("assetsCalled = %v, want %v", assetsCalled, tt.assetsCalled)
			}
		})
	}
}

func TestHandleCreateDocument_WriteDisabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig()) // WriteEnabled = false

	input := CreateDocumentInput{Title: "Test", Content: "Body"}
	result, _, _ := toolkit.handleCreateDocument(t.Context(), nil, input)

	if !result.IsError {
		t.Fatal("expected error when write disabled")
	}
	assertResultContains(t, result, "Write error")
}

func TestHandleUpdateDocument_WriteDisabled(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, DefaultConfig()) // WriteEnabled = false

	input := UpdateDocumentInput{URN: "urn:li:document:doc-1", Title: "Test"}
	result, _, _ := toolkit.handleUpdateDocument(t.Context(), nil, input)

	if !result.IsError {
		t.Fatal("expected error when write disabled")
	}
	assertResultContains(t, result, "Write error")
}

func TestHandleGetEntity_DocumentURN(t *testing.T) {
	mock := &mockClient{}
	mock.getDocumentFunc = func(_ context.Context, urn string) (*types.Document, error) {
		return &types.Document{
			URN:     urn,
			Title:   "My Document",
			Content: "Document body",
			Status:  "PUBLISHED",
		}, nil
	}

	toolkit := NewToolkit(mock, DefaultConfig())
	input := GetEntityInput{URN: "urn:li:document:doc-1"}
	result, _, _ := toolkit.handleGetEntity(t.Context(), nil, input)

	if result.IsError {
		t.Fatal("unexpected error result")
	}
	assertResultContains(t, result, "My Document")
	assertResultContains(t, result, "Document body")
}

func TestRegisterDocumentTools(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	impl := &mcp.Implementation{Name: "test", Version: "1.0.0"}
	server := mcp.NewServer(impl, nil)
	toolkit.RegisterAll(server)

	if !toolkit.registeredTools[ToolCreateDocument] {
		t.Error("RegisterAll should register ToolCreateDocument when WriteEnabled")
	}
	if !toolkit.registeredTools[ToolUpdateDocument] {
		t.Error("RegisterAll should register ToolUpdateDocument when WriteEnabled")
	}
}

// assertResultContains checks that a result's text content contains the expected string.
func assertResultContains(t *testing.T, result *mcp.CallToolResult, expected string) {
	t.Helper()
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			if strings.Contains(tc.Text, expected) {
				return
			}
		}
	}
	t.Errorf("result should contain %q", expected)
}
