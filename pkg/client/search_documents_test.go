package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchDocuments(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		query     string
		wantCount int
		wantErr   bool
	}{
		{
			name:  "documents found",
			query: "incident runbook",
			response: `{"data": {"searchAcrossEntities": {
				"start": 0,
				"count": 10,
				"total": 2,
				"searchResults": [
					{"entity": {
						"urn": "urn:li:document:runbook-1",
						"type": "DOCUMENT",
						"subType": "RUNBOOK",
						"info": {
							"title": "Incident Runbook",
							"contents": {"text": "Step 1: Check logs"},
							"status": {"state": "PUBLISHED"},
							"relatedAssets": [{"asset": {"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)"}}]
						},
						"settings": {"showInGlobalContext": true}
					}},
					{"entity": {
						"urn": "urn:li:document:faq-1",
						"type": "DOCUMENT",
						"subType": "FAQ",
						"info": {
							"title": "Onboarding FAQ",
							"contents": {"text": "Frequently asked questions"},
							"status": {"state": "PUBLISHED"}
						},
						"settings": {"showInGlobalContext": false}
					}}
				]
			}}}`,
			wantCount: 2,
		},
		{
			name:  "list all with wildcard",
			query: "*",
			response: `{"data": {"searchAcrossEntities": {
				"start": 0,
				"count": 10,
				"total": 0,
				"searchResults": []
			}}}`,
			wantCount: 0,
		},
		{
			name:     "graphql error",
			query:    "test",
			response: `{"errors": [{"message": "search failed"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request scopes the search to the DOCUMENT entity type.
				body, _ := io.ReadAll(r.Body)
				var req struct {
					Variables struct {
						Input struct {
							Query string   `json:"query"`
							Types []string `json:"types"`
						} `json:"input"`
					} `json:"variables"`
				}
				_ = json.Unmarshal(body, &req)
				if len(req.Variables.Input.Types) != 1 || req.Variables.Input.Types[0] != "DOCUMENT" {
					t.Errorf("input.types = %v, want [DOCUMENT]", req.Variables.Input.Types)
				}
				if req.Variables.Input.Query != tt.query {
					t.Errorf("input.query = %q, want %q", req.Variables.Input.Query, tt.query)
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()

			c := &Client{
				endpoint:   server.URL,
				token:      "test-token",
				httpClient: server.Client(),
				logger:     NopLogger{},
				config:     DefaultConfig(),
			}

			docs, err := c.SearchDocuments(context.Background(), tt.query)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(docs) != tt.wantCount {
				t.Fatalf("docs len = %d, want %d", len(docs), tt.wantCount)
			}
		})
	}
}

func TestSearchDocuments_ResultFields(t *testing.T) {
	response := `{"data": {"searchAcrossEntities": {
		"start": 0,
		"count": 10,
		"total": 1,
		"searchResults": [
			{"entity": {
				"urn": "urn:li:document:runbook-1",
				"type": "DOCUMENT",
				"subType": "RUNBOOK",
				"info": {
					"title": "Incident Runbook",
					"contents": {"text": "Step 1: Check logs"},
					"status": {"state": "PUBLISHED"},
					"relatedAssets": [
						{"asset": {"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)"}},
						{"asset": {"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.orders,PROD)"}}
					]
				},
				"settings": {"showInGlobalContext": true},
				"ownership": {"owners": [{"owner": {"urn": "urn:li:corpuser:alice", "username": "alice"}, "type": "DATAOWNER"}]},
				"tags": {"tags": [{"tag": {"urn": "urn:li:tag:Production", "name": "Production"}}]},
				"glossaryTerms": {"terms": [{"term": {"urn": "urn:li:glossaryTerm:pii", "properties": {"name": "PII"}}}]},
				"domain": {"domain": {"urn": "urn:li:domain:eng", "properties": {"name": "Engineering"}}}
			}}
		]
	}}}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	docs, err := c.SearchDocuments(context.Background(), "runbook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("docs len = %d, want 1", len(docs))
	}

	doc := docs[0]
	if doc.URN != "urn:li:document:runbook-1" {
		t.Errorf("URN = %q, want urn:li:document:runbook-1", doc.URN)
	}
	if doc.Title != "Incident Runbook" {
		t.Errorf("Title = %q, want Incident Runbook", doc.Title)
	}
	if doc.SubType != "RUNBOOK" {
		t.Errorf("SubType = %q, want RUNBOOK", doc.SubType)
	}
	if doc.Settings == nil || !doc.Settings.ShowInGlobalContext {
		t.Errorf("Settings = %+v, want ShowInGlobalContext=true", doc.Settings)
	}
	if len(doc.RelatedAssets) != 2 {
		t.Fatalf("RelatedAssets len = %d, want 2", len(doc.RelatedAssets))
	}
	if doc.RelatedAssets[0].URN != "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)" {
		t.Errorf("RelatedAssets[0].URN = %q", doc.RelatedAssets[0].URN)
	}

	// Governance metadata must survive document search (not just point lookups),
	// matching the fields returned by GetDocument.
	if len(doc.Owners) != 1 || doc.Owners[0].Name != "alice" {
		t.Errorf("Owners = %+v, want alice", doc.Owners)
	}
	if len(doc.Tags) != 1 || doc.Tags[0].Name != "Production" {
		t.Errorf("Tags = %+v, want Production", doc.Tags)
	}
	if len(doc.GlossaryTerms) != 1 || doc.GlossaryTerms[0].Name != "PII" {
		t.Errorf("GlossaryTerms = %+v, want PII", doc.GlossaryTerms)
	}
	if doc.Domain == nil || doc.Domain.Name != "Engineering" {
		t.Errorf("Domain = %+v, want Engineering", doc.Domain)
	}
}

func TestSearchDocuments_LimitCapping(t *testing.T) {
	var gotCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			Variables struct {
				Input struct {
					Count int `json:"count"`
				} `json:"input"`
			} `json:"variables"`
		}
		_ = json.Unmarshal(body, &req)
		gotCount = req.Variables.Input.Count

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"searchAcrossEntities": {"start": 0, "count": 0, "total": 0, "searchResults": []}}}`))
	}))
	defer server.Close()

	cfg := DefaultConfig()
	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     cfg,
	}

	if _, err := c.SearchDocuments(context.Background(), "*", WithLimit(cfg.MaxLimit+100)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCount != cfg.MaxLimit {
		t.Errorf("count = %d, want capped to MaxLimit %d", gotCount, cfg.MaxLimit)
	}
}
