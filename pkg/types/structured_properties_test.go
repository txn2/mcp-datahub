package types

import (
	"encoding/json"
	"testing"
)

func TestStructuredPropertyDefinition_JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    StructuredPropertyDefinition
		wantJSON string
	}{
		{
			name: "full definition",
			input: StructuredPropertyDefinition{
				URN:           "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
				QualifiedName: "io.acryl.privacy.retentionTime",
				DisplayName:   "Retention Time",
				Description:   "How long data is retained",
				ValueType:     "number",
				Cardinality:   "SINGLE",
				EntityTypes:   []string{"dataset", "dataFlow"},
				AllowedValues: []AllowedValue{
					{Value: "30", Description: "30 days"},
					{Value: "90", Description: "90 days"},
				},
			},
			wantJSON: `{
				"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
				"qualified_name": "io.acryl.privacy.retentionTime",
				"display_name": "Retention Time",
				"description": "How long data is retained",
				"value_type": "number",
				"cardinality": "SINGLE",
				"entity_types": ["dataset", "dataFlow"],
				"allowed_values": [
					{"value": "30", "description": "30 days"},
					{"value": "90", "description": "90 days"}
				]
			}`,
		},
		{
			name: "minimal definition",
			input: StructuredPropertyDefinition{
				URN:           "urn:li:structuredProperty:classification",
				QualifiedName: "classification",
				ValueType:     "string",
			},
			wantJSON: `{
				"urn": "urn:li:structuredProperty:classification",
				"qualified_name": "classification",
				"value_type": "string"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}

			// Unmarshal both to compare as maps (ignoring whitespace)
			var gotMap, wantMap map[string]any
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Fatalf("Unmarshal(got) error: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantMap); err != nil {
				t.Fatalf("Unmarshal(want) error: %v", err)
			}

			// Re-marshal both for stable comparison
			gotNorm, _ := json.Marshal(gotMap)
			wantNorm, _ := json.Marshal(wantMap)
			if string(gotNorm) != string(wantNorm) {
				t.Errorf("Marshal() =\n  %s\nwant:\n  %s", gotNorm, wantNorm)
			}

			// Round-trip: unmarshal back
			var roundTrip StructuredPropertyDefinition
			if err := json.Unmarshal(got, &roundTrip); err != nil {
				t.Fatalf("round-trip Unmarshal() error: %v", err)
			}
			if roundTrip.URN != tt.input.URN {
				t.Errorf("round-trip URN = %q, want %q", roundTrip.URN, tt.input.URN)
			}
			if roundTrip.QualifiedName != tt.input.QualifiedName {
				t.Errorf("round-trip QualifiedName = %q, want %q", roundTrip.QualifiedName, tt.input.QualifiedName)
			}
			if roundTrip.ValueType != tt.input.ValueType {
				t.Errorf("round-trip ValueType = %q, want %q", roundTrip.ValueType, tt.input.ValueType)
			}
		})
	}
}

func TestStructuredPropertyDefinition_UnmarshalJSON(t *testing.T) {
	input := `{
		"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
		"qualified_name": "io.acryl.privacy.retentionTime",
		"display_name": "Retention Time",
		"description": "How long data is retained",
		"value_type": "number",
		"cardinality": "SINGLE",
		"entity_types": ["dataset"],
		"allowed_values": [{"value": "30", "description": "30 days"}]
	}`

	var def StructuredPropertyDefinition
	if err := json.Unmarshal([]byte(input), &def); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if def.URN != "urn:li:structuredProperty:io.acryl.privacy.retentionTime" {
		t.Errorf("URN = %q", def.URN)
	}
	if def.DisplayName != "Retention Time" {
		t.Errorf("DisplayName = %q", def.DisplayName)
	}
	if len(def.EntityTypes) != 1 || def.EntityTypes[0] != "dataset" {
		t.Errorf("EntityTypes = %v", def.EntityTypes)
	}
	if len(def.AllowedValues) != 1 || def.AllowedValues[0].Value != "30" {
		t.Errorf("AllowedValues = %v", def.AllowedValues)
	}
}

func TestStructuredPropertyValue_JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    StructuredPropertyValue
		wantJSON string
	}{
		{
			name: "with definition",
			input: StructuredPropertyValue{
				PropertyURN: "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
				Definition: &StructuredPropertyDefinition{
					URN:           "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
					QualifiedName: "io.acryl.privacy.retentionTime",
					DisplayName:   "Retention Time",
					ValueType:     "number",
				},
				Values: []any{float64(30)},
			},
			wantJSON: `{
				"property_urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
				"definition": {
					"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
					"qualified_name": "io.acryl.privacy.retentionTime",
					"display_name": "Retention Time",
					"value_type": "number"
				},
				"values": [30]
			}`,
		},
		{
			name: "without definition",
			input: StructuredPropertyValue{
				PropertyURN: "urn:li:structuredProperty:classification",
				Values:      []any{"PII", "SENSITIVE"},
			},
			wantJSON: `{
				"property_urn": "urn:li:structuredProperty:classification",
				"values": ["PII", "SENSITIVE"]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}

			var gotMap, wantMap map[string]any
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Fatalf("Unmarshal(got) error: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantJSON), &wantMap); err != nil {
				t.Fatalf("Unmarshal(want) error: %v", err)
			}

			gotNorm, _ := json.Marshal(gotMap)
			wantNorm, _ := json.Marshal(wantMap)
			if string(gotNorm) != string(wantNorm) {
				t.Errorf("Marshal() =\n  %s\nwant:\n  %s", gotNorm, wantNorm)
			}
		})
	}
}

func TestStructuredPropertyValue_UnmarshalJSON(t *testing.T) {
	input := `{
		"property_urn": "urn:li:structuredProperty:classification",
		"values": ["PII", "SENSITIVE"]
	}`

	var val StructuredPropertyValue
	if err := json.Unmarshal([]byte(input), &val); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if val.PropertyURN != "urn:li:structuredProperty:classification" {
		t.Errorf("PropertyURN = %q", val.PropertyURN)
	}
	if len(val.Values) != 2 {
		t.Fatalf("Values len = %d, want 2", len(val.Values))
	}
	if val.Values[0] != "PII" {
		t.Errorf("Values[0] = %v", val.Values[0])
	}
	if val.Values[1] != "SENSITIVE" {
		t.Errorf("Values[1] = %v", val.Values[1])
	}
}

func TestAllowedValue_JSON(t *testing.T) {
	av := AllowedValue{Value: "30", Description: "30 days"}
	got, err := json.Marshal(av)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var roundTrip AllowedValue
	if err := json.Unmarshal(got, &roundTrip); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if roundTrip.Value != "30" {
		t.Errorf("Value = %q, want %q", roundTrip.Value, "30")
	}
	if roundTrip.Description != "30 days" {
		t.Errorf("Description = %q, want %q", roundTrip.Description, "30 days")
	}
}

func TestAllowedValue_OmitEmpty(t *testing.T) {
	av := AllowedValue{Value: "x"}
	got, err := json.Marshal(av)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if _, ok := m["description"]; ok {
		t.Error("description should be omitted when empty")
	}
}
