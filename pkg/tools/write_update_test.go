package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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
		name       string
		input      UpdateInput
		wantWhat   string
		wantAction string
	}{
		{
			"description",
			UpdateInput{What: "description", URN: urn, Value: "new desc"},
			"description (editableDatasetProperties)", "updated",
		},
		{
			"column_description",
			UpdateInput{What: "column_description", URN: urn, FieldPath: "col1", Value: "desc"},
			"column_description", "updated",
		},
		{
			"structured_property",
			UpdateInput{What: "structured_property", URN: urn, Name: "new name"},
			"structured_property", "updated",
		},
		{
			"incident_status",
			UpdateInput{What: "incident_status", URN: "urn:li:incident:1", State: "RESOLVED"},
			"incident_status", "updated to RESOLVED",
		},
		{
			"incident",
			UpdateInput{What: "incident", URN: "urn:li:incident:1", Name: "updated title"},
			"incident", "updated",
		},
		{
			"query",
			UpdateInput{What: "query", URN: "urn:li:query:1", Value: "SELECT 2"},
			"query", "updated",
		},
		{
			"document_contents",
			UpdateInput{What: "document_contents", URN: "urn:li:document:1", Title: "new title"},
			"document_contents", "updated",
		},
		{
			"document_status",
			UpdateInput{What: "document_status", URN: "urn:li:document:1", Value: "PUBLISHED"},
			"document_status", "updated",
		},
		{
			"document_related_entities",
			UpdateInput{What: "document_related_entities", URN: "urn:li:document:1"},
			"document_related_entities", "updated",
		},
		{
			"document_sub_type",
			UpdateInput{What: "document_sub_type", URN: "urn:li:document:1", Value: "FAQ"},
			"document_sub_type", "updated",
		},
		{
			"data_contract",
			UpdateInput{What: "data_contract", URN: urn},
			"data_contract", "updated",
		},
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
			typed, ok := out.(*UpdateOutput)
			if !ok {
				t.Fatal("output should be *UpdateOutput")
			}
			if typed.What != tt.wantWhat {
				t.Errorf("what = %q, want %q", typed.What, tt.wantWhat)
			}
			if typed.Action != tt.wantAction {
				t.Errorf("action = %q, want %q", typed.Action, tt.wantAction)
			}
		})
	}
}

func TestHandleUpdate_MetadataOperations(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	tests := []struct {
		name          string
		input         UpdateInput
		wantWhat      string
		wantAction    string
		wantTargetURN string
	}{
		{
			"tag_add",
			UpdateInput{What: "tag", Action: "add", URN: urn, TargetURN: "urn:li:tag:PII"},
			"tag", "added", "urn:li:tag:PII",
		},
		{
			"tag_remove",
			UpdateInput{What: "tag", Action: "remove", URN: urn, TargetURN: "urn:li:tag:PII"},
			"tag", "removed", "urn:li:tag:PII",
		},
		{
			// Issue #181: a lone value carrying the tag URN is accepted as target.
			"tag_add_via_value",
			UpdateInput{What: "tag", Action: "add", URN: urn, Value: "urn:li:tag:PII"},
			"tag", "added", "urn:li:tag:PII",
		},
		{
			// target_urn and value agree: not a conflict.
			"tag_add_both_equal",
			UpdateInput{What: "tag", Action: "add", URN: urn, TargetURN: "urn:li:tag:PII", Value: "urn:li:tag:PII"},
			"tag", "added", "urn:li:tag:PII",
		},
		{
			"glossary_term_add_via_value",
			UpdateInput{What: "glossary_term", Action: "add", URN: urn, Value: "urn:li:glossaryTerm:X"},
			"glossary_term", "added", "urn:li:glossaryTerm:X",
		},
		{
			"owner_add_via_value",
			UpdateInput{What: "owner", Action: "add", URN: urn, Value: "urn:li:corpuser:user1"},
			"owner", "added", "urn:li:corpuser:user1",
		},
		{
			"domain_set_via_value",
			UpdateInput{What: "domain", Action: "set", URN: urn, Value: "urn:li:domain:eng"},
			"domain", "set", "urn:li:domain:eng",
		},
		{
			"glossary_term_add",
			UpdateInput{What: "glossary_term", Action: "add", URN: urn, TargetURN: "urn:li:glossaryTerm:X"},
			"glossary_term", "added", "urn:li:glossaryTerm:X",
		},
		{
			"glossary_term_remove",
			UpdateInput{What: "glossary_term", Action: "remove", URN: urn, TargetURN: "urn:li:glossaryTerm:X"},
			"glossary_term", "removed", "urn:li:glossaryTerm:X",
		},
		{
			"link_add",
			UpdateInput{What: "link", Action: "add", URN: urn, URL: "https://example.com", Value: "docs"},
			"link", "added", "",
		},
		{
			"link_remove",
			UpdateInput{What: "link", Action: "remove", URN: urn, URL: "https://example.com"},
			"link", "removed", "",
		},
		{
			"owner_add",
			UpdateInput{What: "owner", Action: "add", URN: urn, TargetURN: "urn:li:corpuser:user1"},
			"owner", "added", "urn:li:corpuser:user1",
		},
		{
			"owner_remove",
			UpdateInput{What: "owner", Action: "remove", URN: urn, TargetURN: "urn:li:corpuser:user1"},
			"owner", "removed", "urn:li:corpuser:user1",
		},
		{
			"domain_set",
			UpdateInput{What: "domain", Action: "set", URN: urn, TargetURN: "urn:li:domain:eng"},
			"domain", "set", "urn:li:domain:eng",
		},
		{
			"domain_remove",
			UpdateInput{What: "domain", Action: "remove", URN: urn},
			"domain", "removed", "",
		},
		{
			"domain_default_action",
			UpdateInput{What: "domain", URN: urn, TargetURN: "urn:li:domain:eng"},
			"domain", "set", "urn:li:domain:eng",
		},
		{"structured_properties_set", UpdateInput{
			What: "structured_properties", Action: "set", URN: urn,
			Properties: []types.StructuredPropertyInput{{PropertyURN: "urn:li:sp:test", Values: []any{"v"}}},
		}, "structured_properties", "updated", ""},
		{"structured_properties_remove", UpdateInput{
			What: "structured_properties", Action: "remove", URN: urn,
			PropertyURNs: []string{"urn:li:sp:test"},
		}, "structured_properties", "removed", ""},
		{"structured_properties_default_action", UpdateInput{
			What: "structured_properties", URN: urn,
			Properties: []types.StructuredPropertyInput{{PropertyURN: "urn:li:sp:test", Values: []any{"v"}}},
		}, "structured_properties", "updated", ""},
		{"custom_properties_set", UpdateInput{
			What: "custom_properties", Action: "set", URN: urn,
			CustomProperties: map[string]string{"source_system": "warehouse"},
		}, "custom_properties", "updated", ""},
		{"custom_properties_remove", UpdateInput{
			What: "custom_properties", Action: "remove", URN: urn,
			PropertyKeys: []string{"source_system"},
		}, "custom_properties", "removed", ""},
		{"custom_properties_default_action", UpdateInput{
			What: "custom_properties", URN: urn,
			CustomProperties: map[string]string{"a": "b"},
		}, "custom_properties", "updated", ""},
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
			typed, ok := out.(*UpdateOutput)
			if !ok {
				t.Fatal("output should be *UpdateOutput")
			}
			if typed.What != tt.wantWhat {
				t.Errorf("what = %q, want %q", typed.What, tt.wantWhat)
			}
			if typed.Action != tt.wantAction {
				t.Errorf("action = %q, want %q", typed.Action, tt.wantAction)
			}
			if typed.TargetURN != tt.wantTargetURN {
				t.Errorf("target_urn = %q, want %q", typed.TargetURN, tt.wantTargetURN)
			}
		})
	}
}

func TestHandleUpdate_ActionRequiredForMetadata(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	tests := []struct {
		name  string
		input UpdateInput
	}{
		{"tag_no_action", UpdateInput{What: "tag", URN: urn, TargetURN: "urn:li:tag:PII"}},
		{"glossary_term_no_action", UpdateInput{What: "glossary_term", URN: urn, TargetURN: "urn:li:glossaryTerm:X"}},
		{"link_no_action", UpdateInput{What: "link", URN: urn, URL: "https://example.com"}},
		{"owner_no_action", UpdateInput{What: "owner", URN: urn, TargetURN: "urn:li:corpuser:user1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)
			if !result.IsError {
				t.Errorf("expected error when action is missing for %s", tt.name)
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
		{"incident_status_no_state", UpdateInput{What: "incident_status", URN: "urn:li:incident:1"}},
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
		{"structured_properties_set_empty", UpdateInput{
			What: "structured_properties", Action: "set", URN: urn,
		}},
		{"structured_properties_remove_empty", UpdateInput{
			What: "structured_properties", Action: "remove", URN: urn,
		}},
		{"custom_properties_invalid_action", UpdateInput{
			What: "custom_properties", Action: "add", URN: urn,
			CustomProperties: map[string]string{"a": "b"},
		}},
		{"custom_properties_set_empty", UpdateInput{
			What: "custom_properties", Action: "set", URN: urn,
		}},
		{"custom_properties_remove_empty", UpdateInput{
			What: "custom_properties", Action: "remove", URN: urn,
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

// TestHandleUpdate_TargetURNAndValueConflict verifies that when target_urn and
// value both carry a URN but disagree, the update is rejected rather than
// silently preferring one (issue #181).
func TestHandleUpdate_TargetURNAndValueConflict(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	tests := []struct {
		name  string
		input UpdateInput
	}{
		{"tag", UpdateInput{
			What: "tag", Action: "add", URN: urn,
			TargetURN: "urn:li:tag:PII", Value: "urn:li:tag:Financial",
		}},
		{"glossary_term", UpdateInput{
			What: "glossary_term", Action: "add", URN: urn,
			TargetURN: "urn:li:glossaryTerm:X", Value: "urn:li:glossaryTerm:Y",
		}},
		{"owner", UpdateInput{
			What: "owner", Action: "add", URN: urn,
			TargetURN: "urn:li:corpuser:a", Value: "urn:li:corpuser:b",
		}},
		{"domain", UpdateInput{
			What: "domain", Action: "set", URN: urn,
			TargetURN: "urn:li:domain:eng", Value: "urn:li:domain:sales",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := toolkit.handleUpdate(context.Background(), nil, tt.input)
			if !result.IsError {
				t.Fatalf("expected error for conflicting target_urn/value")
			}
			msg := result.Content[0].(*mcp.TextContent).Text
			if !strings.Contains(msg, "target_urn") || !strings.Contains(msg, "differ") {
				t.Errorf("error = %q, want it to mention target_urn and differ", msg)
			}
		})
	}
}

// TestHandleUpdate_NeitherTargetNorValueNamesTargetURN verifies the missing-target
// error still names target_urn (not value) when both are absent (issue #181).
func TestHandleUpdate_NeitherTargetNorValueNamesTargetURN(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	urn := "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)"

	result, _, _ := toolkit.handleUpdate(context.Background(), nil,
		UpdateInput{What: "tag", Action: "add", URN: urn})
	if !result.IsError {
		t.Fatal("expected error when neither target_urn nor value is set")
	}
	if msg := result.Content[0].(*mcp.TextContent).Text; !strings.Contains(msg, "target_urn") {
		t.Errorf("error = %q, want it to name target_urn", msg)
	}
}

func TestHandleUpdate_ClientErrorPropagation(t *testing.T) {
	clientErr := errors.New("DataHub API error")
	mock := &mockClient{
		addTagFunc: func(_ context.Context, _, _ string) error {
			return clientErr
		},
	}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	result, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "tag", Action: "add",
		URN: "urn:li:dataset:test", TargetURN: "urn:li:tag:PII",
	})
	if !result.IsError {
		t.Fatal("expected error when client returns error")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	if tc.Text == "" {
		t.Error("error message should not be empty")
	}
}

func TestHandleUpdate_CustomPropertiesClientErrorPropagation(t *testing.T) {
	clientErr := errors.New("DataHub API error")
	mock := &mockClient{
		setCustomPropertiesFunc: func(_ context.Context, _ string, _ map[string]string) error {
			return clientErr
		},
		removeCustomPropertiesFunc: func(_ context.Context, _ string, _ []string) error {
			return clientErr
		},
	}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})
	urn := "urn:li:glossaryTerm:revenue"

	setResult, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "custom_properties", Action: "set", URN: urn,
		CustomProperties: map[string]string{"a": "b"},
	})
	if !setResult.IsError {
		t.Error("expected error when SetCustomProperties fails")
	}

	removeResult, _, _ := toolkit.handleUpdate(context.Background(), nil, UpdateInput{
		What: "custom_properties", Action: "remove", URN: urn,
		PropertyKeys: []string{"a"},
	})
	if !removeResult.IsError {
		t.Error("expected error when RemoveCustomProperties fails")
	}
}
