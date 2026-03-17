package tools

import (
	"context"
	"testing"
)

func TestHandleDelete_RequiresWhat(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleDelete(context.Background(), nil, DeleteInput{URN: "urn:li:tag:test"})
	if !result.IsError {
		t.Error("expected error when what is empty")
	}
}

func TestHandleDelete_RequiresURN(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleDelete(context.Background(), nil, DeleteInput{What: "tag"})
	if !result.IsError {
		t.Error("expected error when urn is empty")
	}
}

func TestHandleDelete_RequiresWriteEnabled(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, DefaultConfig())
	result, _, _ := toolkit.handleDelete(context.Background(), nil, DeleteInput{What: "tag", URN: "urn:li:tag:test"})
	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleDelete_AllTypes(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	tests := []struct {
		what string
		urn  string
	}{
		{"query", "urn:li:query:test"},
		{"tag", "urn:li:tag:test"},
		{"domain", "urn:li:domain:test"},
		{"glossary_entity", "urn:li:glossaryTerm:test"},
		{"data_product", "urn:li:dataProduct:test"},
		{"application", "urn:li:application:test"},
		{"document", "urn:li:document:test"},
		{"structured_property", "urn:li:structuredProperty:test"},
	}

	for _, tt := range tests {
		t.Run(tt.what, func(t *testing.T) {
			result, out, _ := toolkit.handleDelete(context.Background(), nil, DeleteInput{What: tt.what, URN: tt.urn})
			if result.IsError {
				t.Errorf("unexpected error for %s", tt.what)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			typed, ok := out.(*DeleteOutput)
			if !ok {
				t.Fatal("output should be *DeleteOutput")
			}
			if typed.What != tt.what {
				t.Errorf("what = %q, want %q", typed.What, tt.what)
			}
			if typed.Action != "deleted" {
				t.Errorf("action = %q, want deleted", typed.Action)
			}
		})
	}
}

func TestHandleDelete_InvalidWhat(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleDelete(context.Background(), nil, DeleteInput{What: "invalid", URN: "urn:li:tag:test"})
	if !result.IsError {
		t.Error("expected error for invalid what value")
	}
}
