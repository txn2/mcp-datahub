package tools

import (
	"context"
	"testing"
)

func TestHandleCreate_RequiresWhat(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleCreate(context.Background(), nil, CreateInput{})
	if !result.IsError {
		t.Error("expected error when what is empty")
	}
}

func TestHandleCreate_RequiresWriteEnabled(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, DefaultConfig())
	result, _, _ := toolkit.handleCreate(context.Background(), nil, CreateInput{What: "tag", Name: "test"})
	if !result.IsError {
		t.Error("expected error when write is disabled")
	}
}

func TestHandleCreate_AllTypes(t *testing.T) {
	mock := &mockClient{}
	toolkit := NewToolkit(mock, Config{WriteEnabled: true})

	tests := []struct {
		name  string
		input CreateInput
	}{
		{"tag", CreateInput{What: "tag", Name: "TestTag"}},
		{"domain", CreateInput{What: "domain", Name: "TestDomain"}},
		{"glossary_term", CreateInput{What: "glossary_term", Name: "TestTerm"}},
		{"data_product", CreateInput{What: "data_product", Name: "TestDP", DomainURN: "urn:li:domain:test"}},
		{"document", CreateInput{What: "document", Name: "TestDoc"}},
		{"application", CreateInput{What: "application", Name: "TestApp"}},
		{"query", CreateInput{What: "query", Value: "SELECT 1"}},
		{"incident", CreateInput{What: "incident", Name: "Down", IncidentType: "OPERATIONAL", EntityURNs: []string{"urn:li:dataset:test"}}},
		{"structured_property", CreateInput{
			What: "structured_property", QualifiedName: "io.test.prop",
			ValueType: "string", EntityTypes: []string{"dataset"},
		}},
		{"data_contract", CreateInput{What: "data_contract", DatasetURNs: []string{"urn:li:dataset:test"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, out, _ := toolkit.handleCreate(context.Background(), nil, tt.input)
			if result.IsError {
				t.Errorf("unexpected error for %s", tt.name)
			}
			if out == nil {
				t.Fatal("expected non-nil output")
			}
			typed, ok := out.(*CreateOutput)
			if !ok {
				t.Fatal("output should be *CreateOutput")
			}
			if typed.What != tt.input.What {
				t.Errorf("what = %q, want %q", typed.What, tt.input.What)
			}
			if typed.Action != "created" {
				t.Errorf("action = %q, want created", typed.Action)
			}
		})
	}
}

func TestHandleCreate_InvalidWhat(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})
	result, _, _ := toolkit.handleCreate(context.Background(), nil, CreateInput{What: "invalid"})
	if !result.IsError {
		t.Error("expected error for invalid what value")
	}
}

func TestHandleCreate_MissingRequiredFields(t *testing.T) {
	toolkit := NewToolkit(&mockClient{}, Config{WriteEnabled: true})

	tests := []struct {
		name  string
		input CreateInput
	}{
		{"tag_no_name", CreateInput{What: "tag"}},
		{"domain_no_name", CreateInput{What: "domain"}},
		{"glossary_term_no_name", CreateInput{What: "glossary_term"}},
		{"data_product_no_name", CreateInput{What: "data_product"}},
		{"data_product_no_domain", CreateInput{What: "data_product", Name: "test"}},
		{"document_no_name", CreateInput{What: "document"}},
		{"application_no_name", CreateInput{What: "application"}},
		{"query_no_value", CreateInput{What: "query"}},
		{"incident_no_urns", CreateInput{What: "incident", Name: "test"}},
		{"incident_no_name", CreateInput{What: "incident", IncidentType: "OPERATIONAL", EntityURNs: []string{"urn:li:dataset:test"}}},
		{"incident_no_type", CreateInput{What: "incident", Name: "Down", EntityURNs: []string{"urn:li:dataset:test"}}},
		{"structured_property_no_qname", CreateInput{What: "structured_property", ValueType: "string", EntityTypes: []string{"dataset"}}},
		{"structured_property_no_vtype", CreateInput{What: "structured_property", QualifiedName: "io.test", EntityTypes: []string{"dataset"}}},
		{"structured_property_no_etypes", CreateInput{What: "structured_property", QualifiedName: "io.test", ValueType: "string"}},
		{"data_contract_no_dataset", CreateInput{What: "data_contract"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, _ := toolkit.handleCreate(context.Background(), nil, tt.input)
			if !result.IsError {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
