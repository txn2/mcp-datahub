package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				URL:   "https://datahub.example.com",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "missing URL",
			config: Config{
				Token: "test-token",
			},
			wantErr: true,
		},
		{
			name: "missing token",
			config: Config{
				URL: "https://datahub.example.com",
			},
			wantErr: true,
		},
		{
			name: "URL without graphql suffix",
			config: Config{
				URL:   "https://datahub.example.com",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "URL with trailing slash",
			config: Config{
				URL:   "https://datahub.example.com/",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "URL already has graphql suffix",
			config: Config{
				URL:   "https://datahub.example.com/api/graphql",
				Token: "test-token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}
			if client == nil {
				t.Error("New() returned nil client")
			}
		})
	}
}

func TestClientDefaults(t *testing.T) {
	cfg := Config{
		URL:   "https://datahub.example.com",
		Token: "test-token",
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	config := client.Config()
	if config.Timeout != 30*time.Second {
		t.Errorf("Default Timeout = %v, want %v", config.Timeout, 30*time.Second)
	}
	if config.RetryMax != 3 {
		t.Errorf("Default RetryMax = %v, want %v", config.RetryMax, 3)
	}
	if config.DefaultLimit != 10 {
		t.Errorf("Default DefaultLimit = %v, want %v", config.DefaultLimit, 10)
	}
	if config.MaxLimit != 100 {
		t.Errorf("Default MaxLimit = %v, want %v", config.MaxLimit, 100)
	}
}

func TestClientClose(t *testing.T) {
	client, err := New(Config{
		URL:   "https://datahub.example.com",
		Token: "test-token",
	})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() unexpected error: %v", err)
	}
}

func TestClientExecute(t *testing.T) {
	tests := []struct {
		name       string
		response   interface{}
		statusCode int
		wantErr    bool
		errType    error
	}{
		{
			name: "successful response",
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"test": "value",
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "unauthorized",
			response:   map[string]interface{}{},
			statusCode: http.StatusUnauthorized,
			wantErr:    true,
			errType:    ErrUnauthorized,
		},
		{
			name:       "forbidden",
			response:   map[string]interface{}{},
			statusCode: http.StatusForbidden,
			wantErr:    true,
			errType:    ErrForbidden,
		},
		{
			name:       "rate limited",
			response:   map[string]interface{}{},
			statusCode: http.StatusTooManyRequests,
			wantErr:    true,
			errType:    ErrRateLimited,
		},
		{
			name: "graphql error",
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": "some graphql error"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "graphql not found error",
			response: map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": "Entity not found"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    true,
			errType:    ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json")
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Expected Authorization header with bearer token")
				}

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client, err := New(Config{
				URL:      server.URL,
				Token:    "test-token",
				RetryMax: 0, // No retries for faster tests
			})
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			var result map[string]interface{}
			err = client.Execute(context.Background(), "query { test }", nil, &result)

			if tt.wantErr {
				if err == nil {
					t.Error("Execute() expected error, got nil")
					return
				}
				if tt.errType != nil && err != tt.errType {
					// Check if it's wrapped
					if !containsErr(err.Error(), tt.errType.Error()) {
						t.Errorf("Execute() error = %v, want %v", err, tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestClientPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"__typename": "Query",
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = client.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping() unexpected error: %v", err)
	}
}

func TestClientSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"search": map[string]interface{}{
					"start": 0,
					"count": 10,
					"total": 1,
					"searchResults": []map[string]interface{}{
						{
							"entity": map[string]interface{}{
								"urn":         "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.table,PROD)",
								"type":        "DATASET",
								"name":        "table",
								"description": "Test table",
								"platform": map[string]interface{}{
									"name": "snowflake",
								},
								"properties": map[string]interface{}{
									"name":        "",
									"description": "",
								},
							},
							"matchedFields": []map[string]interface{}{
								{"name": "name", "value": "table"},
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := client.Search(context.Background(), "table")
	if err != nil {
		t.Errorf("Search() unexpected error: %v", err)
		return
	}

	if result.Total != 1 {
		t.Errorf("Search() Total = %d, want 1", result.Total)
	}
	if len(result.Entities) != 1 {
		t.Errorf("Search() Entities count = %d, want 1", len(result.Entities))
		return
	}
	if result.Entities[0].Name != "table" {
		t.Errorf("Search() Entity name = %s, want table", result.Entities[0].Name)
	}
}

func TestClientSearchOptions(t *testing.T) {
	var receivedInput map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables map[string]interface{} `json:"variables"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		receivedInput = req.Variables["input"].(map[string]interface{})

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"search": map[string]interface{}{
					"start":         0,
					"count":         0,
					"total":         0,
					"searchResults": []interface{}{},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	_, err = client.Search(context.Background(), "test",
		WithEntityType("DASHBOARD"),
		WithLimit(50),
		WithOffset(10),
	)
	if err != nil {
		t.Errorf("Search() unexpected error: %v", err)
		return
	}

	if receivedInput["type"] != "DASHBOARD" {
		t.Errorf("Search() entity type = %v, want DASHBOARD", receivedInput["type"])
	}
	if receivedInput["count"] != float64(50) {
		t.Errorf("Search() count = %v, want 50", receivedInput["count"])
	}
	if receivedInput["start"] != float64(10) {
		t.Errorf("Search() start = %v, want 10", receivedInput["start"])
	}
}

func TestClientGetEntity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"entity": map[string]interface{}{
					"urn":         "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.table,PROD)",
					"type":        "DATASET",
					"name":        "table",
					"description": "Test description",
					"platform": map[string]interface{}{
						"name": "snowflake",
					},
					"properties": map[string]interface{}{
						"name":        "",
						"description": "",
					},
					"subTypes": map[string]interface{}{
						"typeNames": []string{},
					},
					"deprecation": map[string]interface{}{
						"deprecated":       false,
						"note":             "",
						"decommissionTime": 0,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	entity, err := client.GetEntity(context.Background(), "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.schema.table,PROD)")
	if err != nil {
		t.Errorf("GetEntity() unexpected error: %v", err)
		return
	}

	if entity.Name != "table" {
		t.Errorf("GetEntity() Name = %s, want table", entity.Name)
	}
	if entity.Type != "DATASET" {
		t.Errorf("GetEntity() Type = %s, want DATASET", entity.Type)
	}
}

func TestClientGetEntityNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"entity": map[string]interface{}{
					"urn": "",
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	_, err = client.GetEntity(context.Background(), "urn:li:dataset:nonexistent")
	if err == nil {
		t.Error("GetEntity() expected error for not found")
	}
}

func TestClientGetEntityWithPropertiesAndDeprecation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"entity": map[string]interface{}{
					"urn":         "urn:li:dataset:test",
					"type":        "DATASET",
					"name":        "original_name",
					"description": "original_desc",
					"platform": map[string]interface{}{
						"name": "snowflake",
					},
					"properties": map[string]interface{}{
						"name":        "Property Name",
						"description": "Property Description",
					},
					"subTypes": map[string]interface{}{
						"typeNames": []string{},
					},
					"deprecation": map[string]interface{}{
						"deprecated":       true,
						"note":             "This entity is deprecated",
						"decommissionTime": 1704067200000,
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	entity, err := client.GetEntity(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("GetEntity() unexpected error: %v", err)
		return
	}

	// Verify properties override name and description
	if entity.Name != "Property Name" {
		t.Errorf("GetEntity() Name = %s, want Property Name", entity.Name)
	}
	if entity.Description != "Property Description" {
		t.Errorf("GetEntity() Description = %s, want Property Description", entity.Description)
	}

	// Verify deprecation
	if entity.Deprecation == nil {
		t.Error("GetEntity() Deprecation should not be nil")
		return
	}
	if !entity.Deprecation.Deprecated {
		t.Error("GetEntity() Deprecation.Deprecated should be true")
	}
	if entity.Deprecation.Note != "This entity is deprecated" {
		t.Error("GetEntity() Deprecation.Note mismatch")
	}
}

func TestClientGetSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dataset": map[string]interface{}{
					"schemaMetadata": map[string]interface{}{
						"name":        "schema",
						"version":     1,
						"hash":        "abc123",
						"primaryKeys": []string{"id"},
						"fields": []map[string]interface{}{
							{
								"fieldPath":      "id",
								"type":           "NUMBER",
								"nativeDataType": "INT64",
								"description":    "Primary key",
								"nullable":       false,
								"isPartOfKey":    true,
							},
							{
								"fieldPath":      "name",
								"type":           "STRING",
								"nativeDataType": "VARCHAR",
								"description":    "Name field",
								"nullable":       true,
								"isPartOfKey":    false,
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	schema, err := client.GetSchema(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("GetSchema() unexpected error: %v", err)
		return
	}

	if len(schema.Fields) != 2 {
		t.Errorf("GetSchema() Fields count = %d, want 2", len(schema.Fields))
	}
	if schema.Fields[0].FieldPath != "id" {
		t.Errorf("GetSchema() first field = %s, want id", schema.Fields[0].FieldPath)
	}
}

func TestClientGetLineage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"searchAcrossLineage": map[string]interface{}{
					"searchResults": []map[string]interface{}{
						{
							"entity": map[string]interface{}{
								"urn":         "urn:li:dataset:downstream",
								"type":        "DATASET",
								"name":        "downstream_table",
								"description": "Downstream table",
								"platform": map[string]interface{}{
									"name": "snowflake",
								},
							},
							"degree": 1,
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := client.GetLineage(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("GetLineage() unexpected error: %v", err)
		return
	}

	if len(result.Nodes) != 1 {
		t.Errorf("GetLineage() Nodes count = %d, want 1", len(result.Nodes))
	}
}

func TestClientGetQueries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dataset": map[string]interface{}{
					"usageStats": map[string]interface{}{
						"buckets": []map[string]interface{}{
							{
								"metrics": map[string]interface{}{
									"topSqlQueries": []string{
										"SELECT * FROM table",
										"SELECT id FROM table WHERE active = true",
									},
								},
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := client.GetQueries(context.Background(), "urn:li:dataset:test")
	if err != nil {
		t.Errorf("GetQueries() unexpected error: %v", err)
		return
	}

	if result.Total != 2 {
		t.Errorf("GetQueries() Total = %d, want 2", result.Total)
	}
}

func TestClientListTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"search": map[string]interface{}{
					"searchResults": []map[string]interface{}{
						{
							"entity": map[string]interface{}{
								"urn":         "urn:li:tag:PII",
								"name":        "PII",
								"description": "Personal info",
								"properties": map[string]interface{}{
									"name":        "",
									"description": "",
								},
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	tags, err := client.ListTags(context.Background(), "")
	if err != nil {
		t.Errorf("ListTags() unexpected error: %v", err)
		return
	}

	if len(tags) != 1 {
		t.Errorf("ListTags() count = %d, want 1", len(tags))
	}
}

func TestClientListDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"listDomains": map[string]interface{}{
					"total": 1,
					"domains": []map[string]interface{}{
						{
							"urn": "urn:li:domain:marketing",
							"properties": map[string]interface{}{
								"name":        "Marketing",
								"description": "Marketing domain",
							},
							"entities": map[string]interface{}{
								"total": 10,
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	domains, err := client.ListDomains(context.Background())
	if err != nil {
		t.Errorf("ListDomains() unexpected error: %v", err)
		return
	}

	if len(domains) != 1 {
		t.Errorf("ListDomains() count = %d, want 1", len(domains))
	}
	if domains[0].Name != "Marketing" {
		t.Errorf("ListDomains() first domain = %s, want Marketing", domains[0].Name)
	}
}

func TestClientGetGlossaryTerm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"glossaryTerm": map[string]interface{}{
					"urn":              "urn:li:glossaryTerm:business.revenue",
					"name":             "Revenue",
					"hierarchicalName": "Business.Revenue",
					"properties": map[string]interface{}{
						"name":        "Revenue",
						"description": "Total revenue from all sources",
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	term, err := client.GetGlossaryTerm(context.Background(), "urn:li:glossaryTerm:business.revenue")
	if err != nil {
		t.Errorf("GetGlossaryTerm() unexpected error: %v", err)
		return
	}

	if term.Name != "Revenue" {
		t.Errorf("GetGlossaryTerm() Name = %s, want Revenue", term.Name)
	}
	if term.Description != "Total revenue from all sources" {
		t.Errorf("GetGlossaryTerm() Description mismatch")
	}
}

func TestClientGetGlossaryTermNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"glossaryTerm": map[string]interface{}{
					"urn": "",
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	_, err = client.GetGlossaryTerm(context.Background(), "urn:li:glossaryTerm:nonexistent")
	if err == nil {
		t.Error("GetGlossaryTerm() expected error for not found")
	}
}

func TestClientListDataProducts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"listDataProducts": map[string]interface{}{
					"total": 1,
					"dataProducts": []map[string]interface{}{
						{
							"urn": "urn:li:dataProduct:product1",
							"properties": map[string]interface{}{
								"name":        "Product 1",
								"description": "First product",
								"customProperties": []map[string]interface{}{
									{"key": "team", "value": "data-engineering"},
								},
							},
							"domain": map[string]interface{}{
								"domain": map[string]interface{}{
									"urn": "urn:li:domain:marketing",
									"properties": map[string]interface{}{
										"name": "Marketing",
									},
								},
							},
							"ownership": map[string]interface{}{
								"owners": []map[string]interface{}{
									{
										"owner": map[string]interface{}{
											"urn":      "urn:li:corpuser:john",
											"username": "john",
											"name":     "John Doe",
										},
										"type": "TECHNICAL_OWNER",
									},
								},
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	products, err := client.ListDataProducts(context.Background())
	if err != nil {
		t.Errorf("ListDataProducts() unexpected error: %v", err)
		return
	}

	if len(products) != 1 {
		t.Errorf("ListDataProducts() count = %d, want 1", len(products))
		return
	}

	if products[0].Name != "Product 1" {
		t.Errorf("ListDataProducts() first product name = %s, want Product 1", products[0].Name)
	}
	if products[0].Domain == nil || products[0].Domain.Name != "Marketing" {
		t.Error("ListDataProducts() domain not set correctly")
	}
	if len(products[0].Owners) != 1 {
		t.Errorf("ListDataProducts() owners count = %d, want 1", len(products[0].Owners))
	}
	if products[0].Properties["team"] != "data-engineering" {
		t.Error("ListDataProducts() custom properties not set")
	}
}

// TestClientListDataProductsFallback tests the search fallback when listDataProducts query isn't available.
// This is tested via integration tests with a real DataHub instance.
// The direct test is complex because it requires coordinating multiple HTTP calls.

func TestClientGetDataProduct(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dataProduct": map[string]interface{}{
					"urn": "urn:li:dataProduct:test",
					"properties": map[string]interface{}{
						"name":        "Test Product",
						"description": "A test data product",
						"customProperties": []map[string]interface{}{
							{"key": "team", "value": "data-platform"},
						},
					},
					"domain": map[string]interface{}{
						"domain": map[string]interface{}{
							"urn": "urn:li:domain:sales",
							"properties": map[string]interface{}{
								"name":        "Sales",
								"description": "Sales domain",
							},
						},
					},
					"ownership": map[string]interface{}{
						"owners": []map[string]interface{}{
							{
								"owner": map[string]interface{}{
									"urn":      "urn:li:corpuser:jane",
									"username": "jane",
									"name":     "",
									"info": map[string]interface{}{
										"displayName": "Jane Smith",
										"email":       "jane@example.com",
									},
								},
								"type": "DATA_STEWARD",
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	product, err := client.GetDataProduct(context.Background(), "urn:li:dataProduct:test")
	if err != nil {
		t.Errorf("GetDataProduct() unexpected error: %v", err)
		return
	}

	if product.Name != "Test Product" {
		t.Errorf("GetDataProduct() Name = %s, want Test Product", product.Name)
	}
	if product.Domain == nil || product.Domain.Name != "Sales" {
		t.Error("GetDataProduct() domain not set correctly")
	}
	if len(product.Owners) != 1 || product.Owners[0].Name != "Jane Smith" {
		t.Error("GetDataProduct() owners not set correctly")
	}
	if product.Properties["team"] != "data-platform" {
		t.Error("GetDataProduct() custom properties not set")
	}
}

func TestClientGetDataProductNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dataProduct": map[string]interface{}{
					"urn": "",
				},
			},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	_, err = client.GetDataProduct(context.Background(), "urn:li:dataProduct:nonexistent")
	if err == nil {
		t.Error("GetDataProduct() expected error for not found")
	}
}

func TestClientContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{},
		})
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		Timeout:  50 * time.Millisecond,
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx := context.Background()
	err = client.Ping(ctx)
	if err == nil {
		t.Error("Ping() expected timeout error")
	}
}

func TestClientServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client, err := New(Config{
		URL:      server.URL,
		Token:    "test-token",
		RetryMax: 0,
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	err = client.Ping(context.Background())
	if err == nil {
		t.Error("Ping() expected error for server error")
	}
}

func containsErr(got, want string) bool {
	for i := 0; i <= len(got)-len(want); i++ {
		if got[i:i+len(want)] == want {
			return true
		}
	}
	return false
}

func TestNewFromEnv(t *testing.T) {
	// Save existing env vars
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")
	defer func() {
		restoreEnv("DATAHUB_URL", origURL)
		restoreEnv("DATAHUB_TOKEN", origToken)
	}()

	// Set valid env vars
	os.Setenv("DATAHUB_URL", "https://datahub.example.com")
	os.Setenv("DATAHUB_TOKEN", "test-token")

	client, err := NewFromEnv()
	if err != nil {
		t.Errorf("NewFromEnv() unexpected error: %v", err)
		return
	}
	if client == nil {
		t.Error("NewFromEnv() returned nil client")
	}
}

func TestNewFromEnvMissingConfig(t *testing.T) {
	// Save existing env vars
	origURL := os.Getenv("DATAHUB_URL")
	origToken := os.Getenv("DATAHUB_TOKEN")
	defer func() {
		restoreEnv("DATAHUB_URL", origURL)
		restoreEnv("DATAHUB_TOKEN", origToken)
	}()

	// Clear env vars
	os.Unsetenv("DATAHUB_URL")
	os.Unsetenv("DATAHUB_TOKEN")

	_, err := NewFromEnv()
	if err == nil {
		t.Error("NewFromEnv() expected error for missing config")
	}
}

func restoreEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}
