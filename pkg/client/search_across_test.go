package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchAcrossEntities(t *testing.T) {
	tests := []struct {
		name         string
		responseJSON string
		wantTotal    int
		wantCount    int
		wantErr      bool
	}{
		{
			name: "results found",
			responseJSON: `{
				"data": {
					"searchAcrossEntities": {
						"start": 0,
						"count": 10,
						"total": 2,
						"searchResults": [
							{
								"entity": {
									"urn": "urn:li:dataset:(urn:li:dataPlatform:trino,catalog.schema.users,PROD)",
									"type": "DATASET",
									"name": "users",
									"description": "User accounts table",
									"platform": {"name": "trino"}
								},
								"matchedFields": [{"name": "fieldPaths", "value": "email"}]
							},
							{
								"entity": {
									"urn": "urn:li:dataset:(urn:li:dataPlatform:trino,catalog.schema.orders,PROD)",
									"type": "DATASET",
									"name": "orders",
									"description": "Orders table",
									"platform": {"name": "trino"}
								},
								"matchedFields": []
							}
						]
					}
				}
			}`,
			wantTotal: 2,
			wantCount: 2,
		},
		{
			name: "empty results",
			responseJSON: `{
				"data": {
					"searchAcrossEntities": {
						"start": 0,
						"count": 10,
						"total": 0,
						"searchResults": []
					}
				}
			}`,
			wantTotal: 0,
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
				config:     DefaultConfig(),
				logger:     NopLogger{},
			}

			result, err := c.SearchAcrossEntities(context.Background(), "test")
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if result.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", result.Total, tt.wantTotal)
			}
			if len(result.Entities) != tt.wantCount {
				t.Errorf("Entities count = %d, want %d", len(result.Entities), tt.wantCount)
			}
		})
	}
}

func TestSearchAcrossEntities_WithFilters(t *testing.T) {
	var capturedInput map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if variables, ok := req["variables"].(map[string]any); ok {
			capturedInput, _ = variables["input"].(map[string]any)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"searchAcrossEntities": {
					"start": 0, "count": 10, "total": 1,
					"searchResults": [{
						"entity": {
							"urn": "urn:li:dataset:test",
							"type": "DATASET",
							"name": "test"
						},
						"matchedFields": [{"name": "fieldPaths", "value": "email"}]
					}]
				}
			}
		}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	_, err := c.SearchAcrossEntities(context.Background(), "*",
		WithTypes([]string{"DATASET"}),
		WithOrFilters([]SearchFilter{
			{Field: "fieldPaths", Values: []string{"email"}, Condition: "CONTAIN"},
			{Field: "platform", Values: []string{"urn:li:dataPlatform:trino"}},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify types were sent
	types, ok := capturedInput["types"].([]any)
	if !ok || len(types) != 1 || types[0] != "DATASET" {
		t.Errorf("types = %v, want [DATASET]", capturedInput["types"])
	}

	// Verify orFilters were sent
	orFilters, ok := capturedInput["orFilters"].([]any)
	if !ok || len(orFilters) != 1 {
		t.Fatalf("orFilters = %v, want 1 group", capturedInput["orFilters"])
	}

	andGroup, ok := orFilters[0].(map[string]any)
	if !ok {
		t.Fatalf("orFilters[0] not a map")
	}

	andFilters, ok := andGroup["and"].([]any)
	if !ok || len(andFilters) != 2 {
		t.Fatalf("and filters = %v, want 2 filters", andGroup["and"])
	}

	firstFilter, _ := andFilters[0].(map[string]any)
	if firstFilter["field"] != "fieldPaths" {
		t.Errorf("first filter field = %v, want fieldPaths", firstFilter["field"])
	}
	if firstFilter["condition"] != "CONTAIN" {
		t.Errorf("first filter condition = %v, want CONTAIN", firstFilter["condition"])
	}
}

func TestSearchAcrossEntities_WithNegatedFilter(t *testing.T) {
	var capturedInput map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		if variables, ok := req["variables"].(map[string]any); ok {
			capturedInput, _ = variables["input"].(map[string]any)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"searchAcrossEntities": {
					"start": 0, "count": 10, "total": 0, "searchResults": []
				}
			}
		}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	_, err := c.SearchAcrossEntities(context.Background(), "*",
		WithOrFilters([]SearchFilter{
			{Field: "tags", Values: []string{"urn:li:tag:deprecated"}, Negated: true},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	orFilters, ok := capturedInput["orFilters"].([]any)
	if !ok || len(orFilters) != 1 {
		t.Fatalf("orFilters = %v, want 1 group", capturedInput["orFilters"])
	}
	andGroup := orFilters[0].(map[string]any)
	andFilters := andGroup["and"].([]any)
	filter := andFilters[0].(map[string]any)

	if filter["negated"] != true {
		t.Errorf("negated = %v, want true", filter["negated"])
	}
}

func TestSearchAcrossEntities_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "search failed"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	_, err := c.SearchAcrossEntities(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error on GraphQL error response")
	}
}

func TestBuildOrFilters(t *testing.T) {
	filters := []SearchFilter{
		{Field: "fieldPaths", Values: []string{"email", "phone"}, Condition: "CONTAIN"},
		{Field: "platform", Values: []string{"urn:li:dataPlatform:trino"}},
		{Field: "tags", Values: []string{"urn:li:tag:deprecated"}, Negated: true},
	}

	result := buildOrFilters(filters)

	if len(result) != 1 {
		t.Fatalf("expected 1 OR group, got %d", len(result))
	}

	andFilters, ok := result[0]["and"].([]map[string]any)
	if !ok {
		t.Fatal("expected 'and' key with []map[string]any")
	}

	if len(andFilters) != 3 {
		t.Fatalf("expected 3 AND filters, got %d", len(andFilters))
	}

	// First filter: fieldPaths with CONTAIN
	if andFilters[0]["field"] != "fieldPaths" {
		t.Errorf("filter[0].field = %v, want fieldPaths", andFilters[0]["field"])
	}
	if andFilters[0]["condition"] != "CONTAIN" {
		t.Errorf("filter[0].condition = %v, want CONTAIN", andFilters[0]["condition"])
	}

	// Second filter: platform without condition (omitted)
	if _, hasCondition := andFilters[1]["condition"]; hasCondition {
		t.Error("filter[1] should not have condition when empty")
	}

	// Third filter: negated
	if andFilters[2]["negated"] != true {
		t.Errorf("filter[2].negated = %v, want true", andFilters[2]["negated"])
	}

	// Non-negated filters should not have negated key
	if _, hasNegated := andFilters[0]["negated"]; hasNegated {
		t.Error("filter[0] should not have negated key when false")
	}
}

func TestSearchAcrossEntities_EntityDetails(t *testing.T) {
	responseJSON := `{
		"data": {
			"searchAcrossEntities": {
				"start": 0,
				"count": 10,
				"total": 1,
				"searchResults": [{
					"entity": {
						"urn": "urn:li:dataset:(urn:li:dataPlatform:trino,catalog.schema.users,PROD)",
						"type": "DATASET",
						"name": "users",
						"description": "User table",
						"platform": {"name": "trino"},
						"ownership": {
							"owners": [{
								"owner": {"urn": "urn:li:corpuser:admin", "username": "admin"},
								"type": "DATAOWNER"
							}]
						},
						"tags": {
							"tags": [{"tag": {"urn": "urn:li:tag:pii", "name": "pii", "description": "PII data"}}]
						},
						"domain": {
							"domain": {
								"urn": "urn:li:domain:commerce",
								"properties": {"name": "Commerce", "description": "Commerce domain"}
							}
						}
					},
					"matchedFields": [{"name": "fieldPaths", "value": "email"}]
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
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	result, err := c.SearchAcrossEntities(context.Background(), "*",
		WithOrFilters([]SearchFilter{
			{Field: "fieldPaths", Values: []string{"email"}, Condition: "CONTAIN"},
		}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	e := result.Entities[0]
	if e.Platform != "trino" {
		t.Errorf("Platform = %q, want trino", e.Platform)
	}
	if len(e.Owners) != 1 {
		t.Errorf("Owners count = %d, want 1", len(e.Owners))
	}
	if len(e.Tags) != 1 || e.Tags[0].Name != "pii" {
		t.Errorf("Tags = %+v", e.Tags)
	}
	if e.Domain == nil || e.Domain.Name != "Commerce" {
		t.Errorf("Domain = %+v", e.Domain)
	}
	if len(e.MatchedFields) != 1 || e.MatchedFields[0].Name != "fieldPaths" {
		t.Errorf("MatchedFields = %+v", e.MatchedFields)
	}
}
