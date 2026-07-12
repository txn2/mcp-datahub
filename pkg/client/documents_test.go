package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDocument(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name: "full document",
			response: `{"data": {"document": {
				"urn": "urn:li:document:runbook-1",
				"type": "DOCUMENT",
				"subType": "RUNBOOK",
				"info": {
					"title": "Incident Runbook",
					"contents": {"text": "Step 1: Check logs"},
					"source": {"sourceType": "NATIVE"},
					"status": {"state": "PUBLISHED"},
					"created": {"time": 1700000000000},
					"lastModified": {"time": 1700001000000},
					"relatedAssets": [{"asset": {"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)"}}],
					"relatedDocuments": [{"document": {"urn": "urn:li:document:faq-1"}}],
					"parentDocument": {"document": {"urn": "urn:li:document:parent-1"}}
				},
				"settings": {"showInGlobalContext": true},
				"ownership": {"owners": [{"owner": {"urn": "urn:li:corpuser:alice", "username": "alice"}, "type": "DATAOWNER"}]},
				"tags": {"tags": [{"tag": {"urn": "urn:li:tag:uuid-key", "name": "uuid-key", "properties": {"name": "Production"}}}]},
				"glossaryTerms": {"terms": [{"term": {"urn": "urn:li:glossaryTerm:pii", "properties": {"name": "PII"}}}]},
				"domain": {"domain": {"urn": "urn:li:domain:eng", "properties": {"name": "Engineering"}}}
			}}}`,
			wantTitle: "Incident Runbook",
		},
		{
			name:     "not found",
			response: `{"data": {"document": {"urn": ""}}}`,
			wantErr:  true,
		},
		{
			name:     "graphql error",
			response: `{"errors": [{"message": "document not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

			doc, err := c.GetDocument(context.Background(), "urn:li:document:runbook-1")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if doc.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", doc.Title, tt.wantTitle)
			}
			if doc.Content != "Step 1: Check logs" {
				t.Errorf("Content = %q, want %q", doc.Content, "Step 1: Check logs")
			}
			if doc.Status != "PUBLISHED" {
				t.Errorf("Status = %q, want PUBLISHED", doc.Status)
			}
			if doc.SubType != "RUNBOOK" {
				t.Errorf("SubType = %q, want RUNBOOK", doc.SubType)
			}
			if doc.Source == nil || doc.Source.SourceType != "NATIVE" {
				t.Errorf("Source = %+v, want NATIVE", doc.Source)
			}
			if doc.Settings == nil || !doc.Settings.ShowInGlobalContext {
				t.Errorf("Settings = %+v, want ShowInGlobalContext=true", doc.Settings)
			}
			if len(doc.RelatedAssets) != 1 {
				t.Errorf("RelatedAssets len = %d, want 1", len(doc.RelatedAssets))
			}
			if len(doc.RelatedDocuments) != 1 {
				t.Errorf("RelatedDocuments len = %d, want 1", len(doc.RelatedDocuments))
			}
			if doc.ParentDocument == nil || doc.ParentDocument.URN != "urn:li:document:parent-1" {
				t.Errorf("ParentDocument = %+v, want parent-1", doc.ParentDocument)
			}
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
			if doc.Created != 1700000000000 {
				t.Errorf("Created = %d, want 1700000000000", doc.Created)
			}
		})
	}
}

func TestGetRelatedDocuments(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantLen  int
	}{
		{
			name: "with related documents",
			response: `{"data": {"entity": {"relatedDocuments": {
				"total": 2,
				"documents": [
					{"urn": "urn:li:document:doc-1", "info": {"title": "Guide 1", "contents": {"text": "content"}, "status": {"state": "PUBLISHED"}}},
					{"urn": "urn:li:document:doc-2", "info": {"title": "Guide 2", "contents": {"text": "content2"}, "status": {"state": "PUBLISHED"}}}
				]
			}}}}`,
			wantLen: 2,
		},
		{
			name:     "no related documents",
			response: `{"data": {"entity": {}}}`,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

			docs, err := c.GetRelatedDocuments(context.Background(), "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(docs) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(docs), tt.wantLen)
			}
		})
	}
}

// TestGetRelatedDocuments_FullProjection guards that related documents carry the
// same projection as GetDocument/SearchDocuments — specifically settings and
// relatedAssets. Without settings, a hidden document (showInGlobalContext=false)
// parses as Settings==nil and a visibility-gated consumer would surface it.
func TestGetRelatedDocuments_FullProjection(t *testing.T) {
	response := `{"data": {"entity": {"relatedDocuments": {
		"total": 1,
		"documents": [
			{
				"urn": "urn:li:document:hidden-1",
				"subType": "RUNBOOK",
				"info": {
					"title": "Hidden Runbook",
					"contents": {"text": "internal only"},
					"status": {"state": "PUBLISHED"},
					"relatedAssets": [{"asset": {"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)"}}]
				},
				"settings": {"showInGlobalContext": false}
			}
		]
	}}}}`

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

	docs, err := c.GetRelatedDocuments(context.Background(), "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("len = %d, want 1", len(docs))
	}

	doc := docs[0]
	if doc.Settings == nil {
		t.Fatal("Settings == nil; visibility flag did not travel on the related-documents path")
	}
	if doc.Settings.ShowInGlobalContext {
		t.Errorf("ShowInGlobalContext = true, want false (steward hid this document)")
	}
	if len(doc.RelatedAssets) != 1 {
		t.Fatalf("RelatedAssets len = %d, want 1", len(doc.RelatedAssets))
	}
	if doc.RelatedAssets[0].URN != "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)" {
		t.Errorf("RelatedAssets[0].URN = %q", doc.RelatedAssets[0].URN)
	}
}

func TestParseDocumentResponse(t *testing.T) {
	raw := `{
		"urn": "urn:li:document:test",
		"type": "DOCUMENT",
		"subType": "FAQ",
		"info": {
			"title": "My FAQ",
			"contents": {"text": "Q: What?"},
			"status": {"state": "PUBLISHED"},
			"created": {"time": 1700000000000},
			"lastModified": {"time": 1700001000000}
		},
		"settings": {"showInGlobalContext": false}
	}`

	var resp documentResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	doc := parseDocumentResponse(&resp)

	if doc.URN != "urn:li:document:test" {
		t.Errorf("URN = %q", doc.URN)
	}
	if doc.Title != "My FAQ" {
		t.Errorf("Title = %q", doc.Title)
	}
	if doc.Content != "Q: What?" {
		t.Errorf("Content = %q", doc.Content)
	}
	if doc.SubType != "FAQ" {
		t.Errorf("SubType = %q", doc.SubType)
	}
	if doc.Status != "PUBLISHED" {
		t.Errorf("Status = %q", doc.Status)
	}
	if doc.Settings == nil || doc.Settings.ShowInGlobalContext {
		t.Errorf("Settings.ShowInGlobalContext should be false")
	}
	if doc.Created != 1700000000000 {
		t.Errorf("Created = %d", doc.Created)
	}
	if doc.LastModified != 1700001000000 {
		t.Errorf("LastModified = %d", doc.LastModified)
	}
}
