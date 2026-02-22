package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{
			"data": {
				"createQuery": {
					"urn": "urn:li:query:abc123",
					"properties": {
						"name": "Top Revenue",
						"description": "Top revenue query",
						"source": "MANUAL",
						"statement": {
							"value": "SELECT * FROM orders",
							"language": "SQL"
						},
						"created": {
							"time": 1700000000000,
							"actor": "urn:li:corpuser:admin"
						}
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	query, err := c.CreateQuery(context.Background(), CreateQueryInput{
		Name:      "Top Revenue",
		Statement: "SELECT * FROM orders",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query.URN != "urn:li:query:abc123" {
		t.Errorf("expected URN 'urn:li:query:abc123', got %q", query.URN)
	}
	if query.Name != "Top Revenue" {
		t.Errorf("expected Name 'Top Revenue', got %q", query.Name)
	}
	if query.Statement != "SELECT * FROM orders" {
		t.Errorf("expected Statement 'SELECT * FROM orders', got %q", query.Statement)
	}
	if query.Source != "MANUAL" {
		t.Errorf("expected Source 'MANUAL', got %q", query.Source)
	}
	if query.CreatedBy != "urn:li:corpuser:admin" {
		t.Errorf("expected CreatedBy 'urn:li:corpuser:admin', got %q", query.CreatedBy)
	}
}

func TestCreateQuery_EmptyStatement(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	_, err := c.CreateQuery(context.Background(), CreateQueryInput{})
	if err == nil {
		t.Fatal("expected error for empty statement")
	}
}

func TestCreateQuery_WithDatasetURNs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		// Verify subjects are passed as a flat array
		inputRaw, ok := req.Variables["input"].(map[string]any)
		if !ok {
			t.Fatal("expected input variable")
		}
		subjects, ok := inputRaw["subjects"].([]any)
		if !ok {
			t.Fatalf("expected subjects as array in input, got %T", inputRaw["subjects"])
		}
		if len(subjects) != 2 {
			t.Errorf("expected 2 subjects, got %d", len(subjects))
		}
		// Verify each subject has datasetUrn
		for i, s := range subjects {
			subj, ok := s.(map[string]any)
			if !ok {
				t.Fatalf("subject[%d]: expected map, got %T", i, s)
			}
			if _, ok := subj["datasetUrn"]; !ok {
				t.Errorf("subject[%d]: missing datasetUrn field", i)
			}
		}

		resp := `{
			"data": {
				"createQuery": {
					"urn": "urn:li:query:with-datasets",
					"properties": {
						"name": "With Datasets",
						"description": "",
						"source": "MANUAL",
						"statement": {"value": "SELECT 1", "language": "SQL"}
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	query, err := c.CreateQuery(context.Background(), CreateQueryInput{
		Name:      "With Datasets",
		Statement: "SELECT 1",
		DatasetURNs: []string{
			"urn:li:dataset:(urn:li:dataPlatform:hive,db.table1,PROD)",
			"urn:li:dataset:(urn:li:dataPlatform:hive,db.table2,PROD)",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query.URN != "urn:li:query:with-datasets" {
		t.Errorf("unexpected URN: %q", query.URN)
	}
}

func TestCreateQuery_DefaultLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		inputRaw, ok := req.Variables["input"].(map[string]any)
		if !ok {
			t.Fatal("expected input variable")
		}
		props, ok := inputRaw["properties"].(map[string]any)
		if !ok {
			t.Fatal("expected properties in input")
		}
		stmt, ok := props["statement"].(map[string]any)
		if !ok {
			t.Fatal("expected statement in properties")
		}
		if stmt["language"] != "SQL" {
			t.Errorf("expected default language 'SQL', got %v", stmt["language"])
		}

		resp := `{"data":{"createQuery":{"urn":"urn:li:query:lang",` +
			`"properties":{"name":"","description":"","source":"MANUAL",` +
			`"statement":{"value":"SELECT 1","language":"SQL"}}}}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.CreateQuery(context.Background(), CreateQueryInput{
		Statement: "SELECT 1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateQuery_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{"errors":[{"message":"Something went wrong"}]}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.CreateQuery(context.Background(), CreateQueryInput{
		Statement: "SELECT 1",
	})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestUpdateQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{
			"data": {
				"updateQuery": {
					"urn": "urn:li:query:abc123",
					"properties": {
						"name": "Updated Name",
						"description": "Updated desc",
						"source": "MANUAL",
						"statement": {"value": "SELECT 2", "language": "SQL"}
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	query, err := c.UpdateQuery(context.Background(), UpdateQueryInput{
		URN:       "urn:li:query:abc123",
		Name:      "Updated Name",
		Statement: "SELECT 2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query.URN != "urn:li:query:abc123" {
		t.Errorf("expected URN 'urn:li:query:abc123', got %q", query.URN)
	}
	if query.Name != "Updated Name" {
		t.Errorf("expected Name 'Updated Name', got %q", query.Name)
	}
}

func TestUpdateQuery_EmptyURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	_, err := c.UpdateQuery(context.Background(), UpdateQueryInput{})
	if err == nil {
		t.Fatal("expected error for empty URN")
	}
}

func TestUpdateQuery_WithDatasetURNs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		inputRaw, ok := req.Variables["input"].(map[string]any)
		if !ok {
			t.Fatal("expected input variable")
		}
		subjects, ok := inputRaw["subjects"].([]any)
		if !ok {
			t.Fatalf("expected subjects as array in input, got %T", inputRaw["subjects"])
		}
		if len(subjects) != 1 {
			t.Errorf("expected 1 subject, got %d", len(subjects))
		}

		resp := `{"data":{"updateQuery":{"urn":"urn:li:query:abc123",` +
			`"properties":{"name":"","description":"","source":"MANUAL",` +
			`"statement":{"value":"SELECT 1","language":"SQL"}}}}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.UpdateQuery(context.Background(), UpdateQueryInput{
		URN:         "urn:li:query:abc123",
		DatasetURNs: []string{"urn:li:dataset:(urn:li:dataPlatform:hive,db.table1,PROD)"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateQuery_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{"errors":[{"message":"Something went wrong"}]}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	_, err := c.UpdateQuery(context.Background(), UpdateQueryInput{
		URN:       "urn:li:query:abc123",
		Statement: "SELECT 2",
	})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestDeleteQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{"data":{"deleteQuery":true}}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.DeleteQuery(context.Background(), "urn:li:query:abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteQuery_EmptyURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.DeleteQuery(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty URN")
	}
}

func TestDeleteQuery_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := `{"errors":[{"message":"Something went wrong"}]}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.DeleteQuery(context.Background(), "urn:li:query:abc123")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestToQuery_NilProperties(t *testing.T) {
	resp := &queryEntityResponse{URN: "urn:li:query:test"}
	q := toQuery(resp)
	if q.URN != "urn:li:query:test" {
		t.Errorf("expected URN 'urn:li:query:test', got %q", q.URN)
	}
	if q.Name != "" {
		t.Errorf("expected empty Name, got %q", q.Name)
	}
}

func TestToQuery_NilCreated(t *testing.T) {
	resp := &queryEntityResponse{
		URN: "urn:li:query:test",
		Properties: &queryPropertiesRaw{
			Name:   "Test",
			Source: "MANUAL",
			Statement: queryStatementRaw{
				Value:    "SELECT 1",
				Language: "SQL",
			},
		},
	}
	q := toQuery(resp)
	if q.CreatedBy != "" {
		t.Errorf("expected empty CreatedBy, got %q", q.CreatedBy)
	}
	if q.Created != 0 {
		t.Errorf("expected Created 0, got %d", q.Created)
	}
}
