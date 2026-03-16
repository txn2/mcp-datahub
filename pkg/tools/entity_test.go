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

func TestHandleGetEntity_DocumentURN(t *testing.T) {
	tests := []struct {
		name    string
		mockDoc *types.Document
		mockErr error
		wantErr bool
	}{
		{
			name: "successful document get",
			mockDoc: &types.Document{
				URN:     "urn:li:document:runbook-1",
				Title:   "Runbook",
				Content: "Step 1",
				Status:  "PUBLISHED",
			},
			wantErr: false,
		},
		{
			name:    "document not found",
			mockErr: errors.New("not found"),
			wantErr: true,
		},
		{
			name:    "nil document",
			mockDoc: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				getDocumentFunc: func(_ context.Context, _ string) (*types.Document, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return tt.mockDoc, nil
				},
			}

			toolkit := NewToolkit(mock, DefaultConfig())
			result, _, _ := toolkit.handleGetEntity(context.Background(), nil, GetEntityInput{
				URN: "urn:li:document:runbook-1",
			})

			if tt.wantErr {
				if !result.IsError {
					t.Error("expected error result for document URN")
				}
			} else {
				if result.IsError {
					t.Error("unexpected error result for document URN")
				}
			}
		})
	}
}

func TestRelatedDocumentsSupported(t *testing.T) {
	tests := []struct {
		entityType string
		want       bool
	}{
		{"DATASET", true},
		{"GLOSSARY_TERM", true},
		{"GLOSSARY_NODE", true},
		{"CONTAINER", true},
		{"DASHBOARD", false},
		{"CHART", false},
		{"DATA_FLOW", false},
		{"DOCUMENT", false},
	}

	for _, tt := range tests {
		t.Run(tt.entityType, func(t *testing.T) {
			if got := relatedDocumentsSupported(tt.entityType); got != tt.want {
				t.Errorf("relatedDocumentsSupported(%q) = %v, want %v", tt.entityType, got, tt.want)
			}
		})
	}
}
