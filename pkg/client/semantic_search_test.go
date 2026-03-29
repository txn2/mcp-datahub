package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSemanticSearch(t *testing.T) {
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
									"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.users,PROD)",
									"type": "DATASET",
									"name": "users",
									"description": "User accounts table",
									"platform": {"name": "snowflake"},
									"ownership": {
										"owners": [{
											"owner": {"urn": "urn:li:corpuser:admin", "username": "admin"},
											"type": "DATAOWNER"
										}]
									},
									"tags": {
										"tags": [{"tag": {"urn": "urn:li:tag:pii", "name": "pii", "description": ""}}]
									},
									"domain": {
										"domain": {
											"urn": "urn:li:domain:identity",
											"properties": {"name": "Identity", "description": ""}
										}
									}
								},
								"matchedFields": [{"name": "description", "value": "User accounts"}]
							},
							{
								"entity": {
									"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.profiles,PROD)",
									"type": "DATASET",
									"name": "profiles",
									"description": "User profiles",
									"platform": {"name": "snowflake"}
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

			result, err := c.SemanticSearch(context.Background(), "user accounts")
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

func TestSemanticSearch_EntityDetails(t *testing.T) {
	responseJSON := `{
		"data": {
			"searchAcrossEntities": {
				"start": 0,
				"count": 10,
				"total": 1,
				"searchResults": [{
					"entity": {
						"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)",
						"type": "DATASET",
						"name": "users",
						"description": "User table",
						"platform": {"name": "snowflake"},
						"ownership": {
							"owners": [{
								"owner": {"urn": "urn:li:corpuser:admin", "username": "admin", "name": "Admin User"},
								"type": "DATAOWNER"
							}]
						},
						"tags": {
							"tags": [{"tag": {"urn": "urn:li:tag:pii", "name": "pii", "description": "PII data"}}]
						},
						"domain": {
							"domain": {
								"urn": "urn:li:domain:identity",
								"properties": {"name": "Identity", "description": "Identity domain"}
							}
						}
					},
					"matchedFields": [{"name": "description", "value": "User table"}]
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

	result, err := c.SemanticSearch(context.Background(), "user data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := result.Entities[0]
	if e.Platform != "snowflake" {
		t.Errorf("Platform = %q", e.Platform)
	}
	if len(e.Owners) != 1 || e.Owners[0].Name != "Admin User" {
		t.Errorf("Owners = %+v", e.Owners)
	}
	if len(e.Tags) != 1 || e.Tags[0].Name != "pii" {
		t.Errorf("Tags = %+v", e.Tags)
	}
	if e.Domain == nil || e.Domain.Name != "Identity" {
		t.Errorf("Domain = %+v", e.Domain)
	}
	if len(e.MatchedFields) != 1 {
		t.Errorf("MatchedFields count = %d", len(e.MatchedFields))
	}
}

func TestSemanticSearch_GraphQLError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors": [{"message": "semanticSearchAcrossEntities not available"}]}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		config:     DefaultConfig(),
		logger:     NopLogger{},
	}

	// Semantic search should return an error (not empty results) because
	// the caller explicitly requested semantic mode
	_, err := c.SemanticSearch(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error when semantic search is not available")
	}
}

func TestSemanticSearch_DataProductProperties(t *testing.T) {
	responseJSON := `{
		"data": {
			"searchAcrossEntities": {
				"start": 0,
				"count": 10,
				"total": 1,
				"searchResults": [{
					"entity": {
						"urn": "urn:li:dataProduct:product-1",
						"type": "DATA_PRODUCT",
						"properties": {"name": "Analytics Product", "description": "Product description"}
					},
					"matchedFields": []
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

	result, err := c.SemanticSearch(context.Background(), "analytics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := result.Entities[0]
	if e.Name != "Analytics Product" {
		t.Errorf("Name = %q, want %q", e.Name, "Analytics Product")
	}
	if e.Description != "Product description" {
		t.Errorf("Description = %q", e.Description)
	}
}
