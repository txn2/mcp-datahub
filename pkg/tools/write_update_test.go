package tools

import (
	"context"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
)

func TestHandleUpdate_RequiresWhat(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{URN: "urn:li:dataset:test"})
	if !result.IsError {
		t.Error("expected error when what is empty")
	}
}

func TestHandleUpdate_RequiresURN(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{What: "description"})
	if !result.IsError {
		t.Error("expected error when urn is empty")
	}
}

func TestHandleUpdate_RequiresWriteEnabled(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, DefaultConfig())
	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "description", URN: "urn:li:dataset:test", Value: "new desc",
	})
	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleUpdate_AllEntityTypes(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	tests := []struct {
		name  string
		input UpdateInput
	}{
		{"description", UpdateInput{What: "description", URN: urn, Value: "new desc"}},
		{"column_description", UpdateInput{What: "column_description", URN: urn, FieldPath: "col1", Value: "desc"}},
		{"structured_property", UpdateInput{What: "structured_property", URN: urn, Name: "new name"}},
		{"incident_status", UpdateInput{What: "incident_status", URN: "urn:li:incident:1", Value: "resolved"}},
		{"incident", UpdateInput{What: "incident", URN: "urn:li:incident:1", Name: "updated title"}},
		{"query", UpdateInput{What: "query", URN: "urn:li:query:1", Value: "SELECT 2"}},
		{"document_contents", UpdateInput{What: "document_contents", URN: "urn:li:document:1", Title: "new title"}},
		{"document_status", UpdateInput{What: "document_status", URN: "urn:li:document:1", Value: "PUBLISHED"}},
		{"document_related_entities", UpdateInput{What: "document_related_entities", URN: "urn:li:document:1"}},
		{"document_sub_type", UpdateInput{What: "document_sub_type", URN: "urn:li:document:1", Value: "FAQ"}},
		{"data_contract", UpdateInput{What: "data_contract", URN: urn}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, out, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)
			if result.IsError {
				t.Errorf("unexpected error for %s", tt.name)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
		})
	}
}

func TestHandleUpdate_MetadataOperations(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	tests := []struct {
		name  string
		input UpdateInput
	}{
		{"tag_add", UpdateInput{What: "tag", Action: "add", URN: urn, TargetURN: "urn:li:tag:PII"}},
		{"tag_remove", UpdateInput{What: "tag", Action: "remove", URN: urn, TargetURN: "urn:li:tag:PII"}},
		{"glossary_term_add", UpdateInput{What: "glossary_term", Action: "add", URN: urn, TargetURN: "urn:li:glossaryTerm:X"}},
		{"glossary_term_remove", UpdateInput{What: "glossary_term", Action: "remove", URN: urn, TargetURN: "urn:li:glossaryTerm:X"}},
		{"link_add", UpdateInput{What: "link", Action: "add", URN: urn, URL: "https://example.com", Value: "docs"}},
		{"link_remove", UpdateInput{What: "link", Action: "remove", URN: urn, URL: "https://example.com"}},
		{"owner_add", UpdateInput{What: "owner", Action: "add", URN: urn, TargetURN: "urn:li:corpuser:user1"}},
		{"owner_remove", UpdateInput{What: "owner", Action: "remove", URN: urn, TargetURN: "urn:li:corpuser:user1"}},
		{"domain_set", UpdateInput{What: "domain", Action: "set", URN: urn, TargetURN: "urn:li:domain:eng"}},
		{"domain_remove", UpdateInput{What: "domain", Action: "remove", URN: urn}},
		{"structured_properties_set", UpdateInput{
			What: "structured_properties", Action: "set", URN: urn,
			Properties: []types.StructuredPropertyInput{{PropertyURN: "urn:li:sp:test", Values: []any{"v"}}},
		}},
		{"structured_properties_remove", UpdateInput{
			What: "structured_properties", Action: "remove", URN: urn,
			PropertyURNs: []string{"urn:li:sp:test"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)
			if result.IsError {
				t.Errorf("unexpected error for %s", tt.name)
			}
		})
	}
}

func TestHandleUpdate_InvalidWhat(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "invalid", URN: "urn:li:dataset:test",
	})
	if !result.IsError {
		t.Error("expected error for invalid what value")
	}
}

func TestHandleUpdate_MissingRequiredFields(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	tests := []struct {
		name  string
		input UpdateInput
	}{
		{"column_desc_no_field_path", UpdateInput{What: "column_description", URN: urn}},
		{"tag_no_target", UpdateInput{What: "tag", Action: "add", URN: urn}},
		{"glossary_term_no_target", UpdateInput{What: "glossary_term", Action: "add", URN: urn}},
		{"link_no_url", UpdateInput{What: "link", Action: "add", URN: urn}},
		{"owner_no_target", UpdateInput{What: "owner", Action: "add", URN: urn}},
		{"domain_set_no_target", UpdateInput{What: "domain", Action: "set", URN: urn}},
		{"document_status_no_value", UpdateInput{What: "document_status", URN: urn}},
		{"document_sub_type_no_value", UpdateInput{What: "document_sub_type", URN: urn}},
		{"tag_invalid_action", UpdateInput{What: "tag", Action: "set", URN: urn, TargetURN: "urn:li:tag:PII"}},
		{"glossary_term_invalid_action", UpdateInput{What: "glossary_term", Action: "set", URN: urn, TargetURN: "urn:li:glossaryTerm:X"}},
		{"link_invalid_action", UpdateInput{What: "link", Action: "set", URN: urn, URL: "https://example.com"}},
		{"owner_invalid_action", UpdateInput{What: "owner", Action: "set", URN: urn, TargetURN: "urn:li:corpuser:user1"}},
		{"domain_invalid_action", UpdateInput{What: "domain", Action: "add", URN: urn, TargetURN: "urn:li:domain:eng"}},
		{"structured_properties_invalid_action", UpdateInput{
			What: "structured_properties", Action: "add", URN: urn,
			Properties: []types.StructuredPropertyInput{{PropertyURN: "urn:li:sp:test", Values: []any{"v"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)
			if !result.IsError {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
