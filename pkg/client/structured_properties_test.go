package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStructuredProperties(t *testing.T) {
	tests := []struct {
		name          string
		responseJSON  string
		wantCount     int
		wantFirstURN  string
		wantFirstVals int
		wantErr       bool
	}{
		{
			name: "single property with string values",
			responseJSON: `{
				"data": {
					"entity": {
						"structuredProperties": {
							"properties": [{
								"structuredProperty": {
									"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
									"definition": {
										"qualifiedName": "io.acryl.privacy.retentionTime",
										"displayName": "Retention Time",
										"description": "Data retention period",
										"valueType": {"info": {"type": "NUMBER"}},
										"cardinality": "SINGLE",
										"entityTypes": [{"info": {"type": "dataset"}}],
										"allowedValues": [
											{"value": {"numberValue": 30}, "description": "30 days"},
											{"value": {"numberValue": 90}, "description": "90 days"}
										]
									}
								},
								"values": [{"numberValue": 30}]
							}]
						}
					}
				}
			}`,
			wantCount:     1,
			wantFirstURN:  "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
			wantFirstVals: 1,
		},
		{
			name: "multiple properties",
			responseJSON: `{
				"data": {
					"entity": {
						"structuredProperties": {
							"properties": [
								{
									"structuredProperty": {
										"urn": "urn:li:structuredProperty:classification",
										"definition": {
											"qualifiedName": "classification",
											"displayName": "Classification",
											"description": "",
											"valueType": {"info": {"type": "STRING"}},
											"cardinality": "MULTIPLE",
											"entityTypes": [],
											"allowedValues": []
										}
									},
									"values": [
										{"stringValue": "PII"},
										{"stringValue": "SENSITIVE"}
									]
								},
								{
									"structuredProperty": {
										"urn": "urn:li:structuredProperty:sla",
										"definition": null
									},
									"values": [{"stringValue": "tier-1"}]
								}
							]
						}
					}
				}
			}`,
			wantCount:     2,
			wantFirstURN:  "urn:li:structuredProperty:classification",
			wantFirstVals: 2,
		},
		{
			name: "no structured properties aspect",
			responseJSON: `{
				"data": {
					"entity": {
						"structuredProperties": null
					}
				}
			}`,
			wantCount: 0,
		},
		{
			name: "empty properties list",
			responseJSON: `{
				"data": {
					"entity": {
						"structuredProperties": {
							"properties": []
						}
					}
				}
			}`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.responseJSON))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			result, err := c.GetStructuredProperties(context.Background(), "urn:li:dataset:test")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if len(result) != tt.wantCount {
				t.Fatalf("got %d properties, want %d", len(result), tt.wantCount)
			}

			if tt.wantCount > 0 {
				if result[0].PropertyURN != tt.wantFirstURN {
					t.Errorf("first property URN = %q, want %q", result[0].PropertyURN, tt.wantFirstURN)
				}
				if len(result[0].Values) != tt.wantFirstVals {
					t.Errorf("first property values count = %d, want %d", len(result[0].Values), tt.wantFirstVals)
				}
			}
		})
	}
}

func TestGetStructuredProperties_Definition(t *testing.T) {
	responseJSON := `{
		"data": {
			"entity": {
				"structuredProperties": {
					"properties": [{
						"structuredProperty": {
							"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
							"definition": {
								"qualifiedName": "io.acryl.privacy.retentionTime",
								"displayName": "Retention Time",
								"description": "How long data is retained",
								"valueType": {"info": {"type": "NUMBER"}},
								"cardinality": "SINGLE",
								"entityTypes": [{"info": {"type": "dataset"}}, {"info": {"type": "dataFlow"}}],
								"allowedValues": [
									{"value": {"numberValue": 30}, "description": "30 days"},
									{"value": {"numberValue": 90}, "description": "90 days"}
								]
							}
						},
						"values": [{"numberValue": 30}]
					}]
				}
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	result, err := c.GetStructuredProperties(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("got %d properties, want 1", len(result))
	}

	prop := result[0]
	if prop.Definition == nil {
		t.Fatal("expected definition to be populated")
	}
	if prop.Definition.QualifiedName != "io.acryl.privacy.retentionTime" {
		t.Errorf("QualifiedName = %q", prop.Definition.QualifiedName)
	}
	if prop.Definition.DisplayName != "Retention Time" {
		t.Errorf("DisplayName = %q", prop.Definition.DisplayName)
	}
	if prop.Definition.ValueType != "NUMBER" {
		t.Errorf("ValueType = %q", prop.Definition.ValueType)
	}
	if prop.Definition.Cardinality != "SINGLE" {
		t.Errorf("Cardinality = %q", prop.Definition.Cardinality)
	}
	if len(prop.Definition.EntityTypes) != 2 {
		t.Errorf("EntityTypes count = %d, want 2", len(prop.Definition.EntityTypes))
	}
	if len(prop.Definition.AllowedValues) != 2 {
		t.Errorf("AllowedValues count = %d, want 2", len(prop.Definition.AllowedValues))
	}
	if prop.Definition.AllowedValues[0].Value != "30" {
		t.Errorf("AllowedValues[0].Value = %q, want %q", prop.Definition.AllowedValues[0].Value, "30")
	}

	// Check the actual value
	if len(prop.Values) != 1 {
		t.Fatalf("Values count = %d, want 1", len(prop.Values))
	}
	if prop.Values[0] != float64(30) {
		t.Errorf("Values[0] = %v (type %T), want 30", prop.Values[0], prop.Values[0])
	}
}

func TestGetStructuredProperties_GraphQLError_ReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "entity not found"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	// Should return empty results (not error) for backward compatibility with DataHub < 1.4.x
	result, err := c.GetStructuredProperties(context.Background(), "urn:li:dataset:nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for graceful degradation, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestListStructuredPropertyDefinitions(t *testing.T) {
	tests := []struct {
		name         string
		responseJSON string
		wantCount    int
		wantErr      bool
	}{
		{
			name: "multiple definitions",
			responseJSON: `{
				"data": {
					"search": {
						"total": 2,
						"searchResults": [
							{
								"entity": {
									"urn": "urn:li:structuredProperty:io.acryl.privacy.retentionTime",
									"definition": {
										"qualifiedName": "io.acryl.privacy.retentionTime",
										"displayName": "Retention Time",
										"description": "Data retention",
										"valueType": {"info": {"type": "NUMBER"}},
										"cardinality": "SINGLE",
										"entityTypes": [{"info": {"type": "dataset"}}],
										"allowedValues": []
									}
								}
							},
							{
								"entity": {
									"urn": "urn:li:structuredProperty:classification",
									"definition": {
										"qualifiedName": "classification",
										"displayName": "Classification",
										"description": "Data classification",
										"valueType": {"info": {"type": "STRING"}},
										"cardinality": "MULTIPLE",
										"entityTypes": [],
										"allowedValues": [
											{"value": {"stringValue": "PII"}, "description": "Personal data"},
											{"value": {"stringValue": "PUBLIC"}, "description": "Public data"}
										]
									}
								}
							}
						]
					}
				}
			}`,
			wantCount: 2,
		},
		{
			name: "empty results",
			responseJSON: `{
				"data": {
					"search": {
						"total": 0,
						"searchResults": []
					}
				}
			}`,
			wantCount: 0,
		},
		{
			name: "definition with null definition field",
			responseJSON: `{
				"data": {
					"search": {
						"total": 1,
						"searchResults": [{
							"entity": {
								"urn": "urn:li:structuredProperty:orphan",
								"definition": null
							}
						}]
					}
				}
			}`,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.responseJSON))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				config:     Config{MaxLimit: 100},
				logger:     NopLogger{},
			}

			result, err := c.ListStructuredPropertyDefinitions(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("got %d definitions, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestListStructuredPropertyDefinitions_Details(t *testing.T) {
	responseJSON := `{
		"data": {
			"search": {
				"total": 1,
				"searchResults": [{
					"entity": {
						"urn": "urn:li:structuredProperty:classification",
						"definition": {
							"qualifiedName": "classification",
							"displayName": "Data Classification",
							"description": "Classify data sensitivity",
							"valueType": {"info": {"type": "STRING"}},
							"cardinality": "MULTIPLE",
							"entityTypes": [{"info": {"type": "dataset"}}, {"info": {"type": "dataFlow"}}],
							"allowedValues": [
								{"value": {"stringValue": "PII"}, "description": "Personal data"},
								{"value": {"stringValue": "PUBLIC"}, "description": ""}
							]
						}
					}
				}]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     Config{MaxLimit: 100},
		logger:     NopLogger{},
	}

	defs, err := c.ListStructuredPropertyDefinitions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("got %d definitions, want 1", len(defs))
	}

	def := defs[0]
	if def.URN != "urn:li:structuredProperty:classification" {
		t.Errorf("URN = %q", def.URN)
	}
	if def.QualifiedName != "classification" {
		t.Errorf("QualifiedName = %q", def.QualifiedName)
	}
	if def.DisplayName != "Data Classification" {
		t.Errorf("DisplayName = %q", def.DisplayName)
	}
	if def.ValueType != "STRING" {
		t.Errorf("ValueType = %q", def.ValueType)
	}
	if def.Cardinality != "MULTIPLE" {
		t.Errorf("Cardinality = %q", def.Cardinality)
	}
	if len(def.EntityTypes) != 2 {
		t.Errorf("EntityTypes count = %d, want 2", len(def.EntityTypes))
	}
	if len(def.AllowedValues) != 2 {
		t.Errorf("AllowedValues count = %d, want 2", len(def.AllowedValues))
	}
	if def.AllowedValues[0].Value != "PII" {
		t.Errorf("AllowedValues[0].Value = %q", def.AllowedValues[0].Value)
	}
	if def.AllowedValues[0].Description != "Personal data" {
		t.Errorf("AllowedValues[0].Description = %q", def.AllowedValues[0].Description)
	}
}

func TestListStructuredPropertyDefinitions_GraphQLError_ReturnsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "STRUCTURED_PROPERTY type not supported"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     Config{MaxLimit: 100},
		logger:     NopLogger{},
	}

	// Should return empty results (not error) for backward compatibility with DataHub < 1.4.x
	result, err := c.ListStructuredPropertyDefinitions(context.Background())
	if err != nil {
		t.Fatalf("expected nil error for graceful degradation, got: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestUpsertStructuredProperties(t *testing.T) {
	tests := []struct {
		name       string
		urn        string
		properties []StructuredPropertyInput
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name: "single property",
			urn:  "urn:li:dataset:test",
			properties: []StructuredPropertyInput{
				{
					PropertyURN: "urn:li:structuredProperty:retentionTime",
					Values:      []any{float64(30)},
				},
			},
			response: `{"data": {"upsertStructuredProperties": true}}`,
		},
		{
			name: "multiple properties",
			urn:  "urn:li:dataset:test",
			properties: []StructuredPropertyInput{
				{
					PropertyURN: "urn:li:structuredProperty:retentionTime",
					Values:      []any{float64(30)},
				},
				{
					PropertyURN: "urn:li:structuredProperty:classification",
					Values:      []any{"PII", "SENSITIVE"},
				},
			},
			response: `{"data": {"upsertStructuredProperties": true}}`,
		},
		{
			name: "graphql error",
			urn:  "urn:li:dataset:test",
			properties: []StructuredPropertyInput{
				{PropertyURN: "urn:li:structuredProperty:invalid", Values: []any{"x"}},
			},
			response: `{"errors": [{"message": "invalid property URN"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&receivedBody)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			err := c.UpsertStructuredProperties(context.Background(), tt.urn, tt.properties)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && receivedBody != nil {
				// Verify the variables contain the input
				vars, ok := receivedBody["variables"].(map[string]any)
				if !ok {
					t.Fatal("expected variables in request")
				}
				input, ok := vars["input"].(map[string]any)
				if !ok {
					t.Fatal("expected input in variables")
				}
				if input["assetUrn"] != tt.urn {
					t.Errorf("assetUrn = %v, want %v", input["assetUrn"], tt.urn)
				}
			}
		})
	}
}

func TestRemoveStructuredProperties(t *testing.T) {
	tests := []struct {
		name         string
		urn          string
		propertyURNs []string
		response     string
		wantErr      bool
	}{
		{
			name:         "single property",
			urn:          "urn:li:dataset:test",
			propertyURNs: []string{"urn:li:structuredProperty:retentionTime"},
			response:     `{"data": {"removeStructuredProperties": true}}`,
		},
		{
			name: "multiple properties",
			urn:  "urn:li:dataset:test",
			propertyURNs: []string{
				"urn:li:structuredProperty:retentionTime",
				"urn:li:structuredProperty:classification",
			},
			response: `{"data": {"removeStructuredProperties": true}}`,
		},
		{
			name:         "graphql error",
			urn:          "urn:li:dataset:test",
			propertyURNs: []string{"urn:li:structuredProperty:nonexistent"},
			response:     `{"errors": [{"message": "property not found"}]}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewDecoder(r.Body).Decode(&receivedBody)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
			}

			err := c.RemoveStructuredProperties(context.Background(), tt.urn, tt.propertyURNs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && receivedBody != nil {
				vars, ok := receivedBody["variables"].(map[string]any)
				if !ok {
					t.Fatal("expected variables in request")
				}
				input, ok := vars["input"].(map[string]any)
				if !ok {
					t.Fatal("expected input in variables")
				}
				if input["assetUrn"] != tt.urn {
					t.Errorf("assetUrn = %v, want %v", input["assetUrn"], tt.urn)
				}
			}
		})
	}
}

func TestPropertyValueEntry_ToAny(t *testing.T) {
	strVal := "hello"
	numVal := float64(42)

	tests := []struct {
		name  string
		entry propertyValueEntry
		want  any
	}{
		{
			name:  "string value",
			entry: propertyValueEntry{StringValue: &strVal},
			want:  "hello",
		},
		{
			name:  "number value",
			entry: propertyValueEntry{NumberValue: &numVal},
			want:  float64(42),
		},
		{
			name:  "nil value",
			entry: propertyValueEntry{},
			want:  nil,
		},
		{
			name:  "string takes precedence",
			entry: propertyValueEntry{StringValue: &strVal, NumberValue: &numVal},
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.toAny()
			if got != tt.want {
				t.Errorf("toAny() = %v (type %T), want %v (type %T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestParseDefinition_NilEntry(t *testing.T) {
	def := parseDefinition("urn:li:structuredProperty:test", nil)
	if def.URN != "urn:li:structuredProperty:test" {
		t.Errorf("URN = %q", def.URN)
	}
	if def.QualifiedName != "" {
		t.Errorf("QualifiedName should be empty, got %q", def.QualifiedName)
	}
}

func TestParseDefinition_EmptyEntityType(t *testing.T) {
	entry := &structuredPropertyDefEntry{
		QualifiedName: "test",
		EntityTypes: []struct {
			Info struct {
				Type string `json:"type"`
			} `json:"info"`
		}{
			{Info: struct {
				Type string `json:"type"`
			}{Type: ""}},
			{Info: struct {
				Type string `json:"type"`
			}{Type: "dataset"}},
		},
	}

	def := parseDefinition("urn:test", entry)
	if len(def.EntityTypes) != 1 {
		t.Errorf("EntityTypes count = %d, want 1 (empty should be filtered)", len(def.EntityTypes))
	}
	if def.EntityTypes[0] != "dataset" {
		t.Errorf("EntityTypes[0] = %q, want %q", def.EntityTypes[0], "dataset")
	}
}
