package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/txn2/mcp-datahub/pkg/types"
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
					"relatedAssets": [{"urn": "urn:li:dataset:(urn:li:dataPlatform:snowflake,db.users,PROD)"}],
					"relatedDocuments": [{"urn": "urn:li:document:faq-1"}],
					"parentDocument": {"urn": "urn:li:document:parent-1"}
				},
				"settings": {"showInGlobalContext": true},
				"ownership": {"owners": [{"owner": {"urn": "urn:li:corpuser:alice", "username": "alice"}, "type": "DATAOWNER"}]},
				"tags": {"tags": [{"tag": {"urn": "urn:li:tag:Production", "name": "Production"}}]},
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

func TestSearchDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"searchDocuments": {
			"start": 0, "count": 10, "total": 1,
			"searchResults": [{
				"document": {
					"urn": "urn:li:document:doc-1",
					"subType": "FAQ",
					"info": {
						"title": "Data FAQ",
						"contents": {"text": "Q: What is this?"},
						"status": {"state": "PUBLISHED"},
						"created": {"time": 1700000000000}
					},
					"settings": {"showInGlobalContext": true}
				}
			}]
		}}}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	result, err := c.SearchDocuments(context.Background(), "FAQ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if len(result.Documents) != 1 {
		t.Fatalf("Documents len = %d, want 1", len(result.Documents))
	}
	if result.Documents[0].Title != "Data FAQ" {
		t.Errorf("Title = %q, want %q", result.Documents[0].Title, "Data FAQ")
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

func TestCreateDocument(t *testing.T) {
	tests := []struct {
		name       string
		input      types.CreateDocumentInput
		response   string
		wantURN    string
		wantErr    bool
		checkInput func(t *testing.T, input map[string]any)
	}{
		{
			name: "basic document",
			input: types.CreateDocumentInput{
				Title:   "Test Doc",
				Content: "Hello world",
			},
			response: `{"data": {"createDocument": "urn:li:document:new-1"}}`,
			wantURN:  "urn:li:document:new-1",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["title"] != "Test Doc" {
					t.Errorf("title = %v, want Test Doc", input["title"])
				}
				contents, ok := input["contents"].(map[string]any)
				if !ok {
					t.Fatal("contents should be a map")
				}
				if contents["text"] != "Hello world" {
					t.Errorf("contents.text = %v, want Hello world", contents["text"])
				}
			},
		},
		{
			name: "with related assets and settings",
			input: types.CreateDocumentInput{
				Title:               "Guide",
				Content:             "Content",
				State:               "PUBLISHED",
				ShowInGlobalContext: boolPtr(false),
				RelatedAssets:       []string{"urn:li:dataset:test"},
			},
			response: `{"data": {"createDocument": "urn:li:document:new-2"}}`,
			wantURN:  "urn:li:document:new-2",
			checkInput: func(t *testing.T, input map[string]any) {
				t.Helper()
				if input["state"] != "PUBLISHED" {
					t.Errorf("state = %v, want PUBLISHED", input["state"])
				}
				settings, ok := input["settings"].(map[string]any)
				if !ok {
					t.Fatal("settings should be a map")
				}
				if settings["showInGlobalContext"] != false {
					t.Errorf("showInGlobalContext = %v, want false", settings["showInGlobalContext"])
				}
				assets, ok := input["relatedAssets"].([]any)
				if !ok || len(assets) != 1 {
					t.Fatalf("relatedAssets should have 1 element, got %v", input["relatedAssets"])
				}
			},
		},
		{
			name:     "graphql error",
			input:    types.CreateDocumentInput{Title: "Test", Content: "Test"},
			response: `{"errors": [{"message": "permission denied"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkInput != nil {
					input := extractGraphQLInput(t, r)
					tt.checkInput(t, input)
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

			urn, err := c.CreateDocument(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if urn != tt.wantURN {
				t.Errorf("URN = %q, want %q", urn, tt.wantURN)
			}
		})
	}
}

func TestUpdateDocumentContents(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		response string
		wantErr  bool
	}{
		{
			name:     "update both",
			title:    "New Title",
			content:  "New Content",
			response: `{"data": {"updateDocumentContents": true}}`,
		},
		{
			name:     "update title only",
			title:    "New Title",
			response: `{"data": {"updateDocumentContents": true}}`,
		},
		{
			name:     "graphql error",
			title:    "Fail",
			response: `{"errors": [{"message": "not found"}]}`,
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

			err := c.UpdateDocumentContents(context.Background(), "urn:li:document:doc-1", tt.title, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestUpdateDocumentRelatedEntities(t *testing.T) {
	var capturedInput map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInput = extractGraphQLInput(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"updateDocumentRelatedEntities": true}}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	err := c.UpdateDocumentRelatedEntities(context.Background(), "urn:li:document:doc-1", []string{"urn:li:dataset:ds1", "urn:li:dataset:ds2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedInput["urn"] != "urn:li:document:doc-1" {
		t.Errorf("urn = %v, want urn:li:document:doc-1", capturedInput["urn"])
	}

	assets, ok := capturedInput["relatedAssets"].([]any)
	if !ok || len(assets) != 2 {
		t.Fatalf("relatedAssets should have 2 elements, got %v", capturedInput["relatedAssets"])
	}
}

func TestUpdateDocumentStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"updateDocumentStatus": true}}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	err := c.UpdateDocumentStatus(context.Background(), "urn:li:document:doc-1", "UNPUBLISHED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDocumentSettings(t *testing.T) {
	var capturedInput map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInput = extractGraphQLInput(t, r)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data": {"updateDocumentSettings": true}}`))
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL,
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
		config:     DefaultConfig(),
	}

	err := c.UpdateDocumentSettings(context.Background(), "urn:li:document:doc-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedInput["urn"] != "urn:li:document:doc-1" {
		t.Errorf("urn = %v, want urn:li:document:doc-1", capturedInput["urn"])
	}
	if capturedInput["showInGlobalContext"] != false {
		t.Errorf("showInGlobalContext = %v, want false", capturedInput["showInGlobalContext"])
	}
}

func TestDeleteDocument(t *testing.T) {
	tests := []struct {
		name     string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			response: `{"data": {"deleteDocument": true}}`,
		},
		{
			name:     "error",
			response: `{"errors": [{"message": "not found"}]}`,
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

			err := c.DeleteDocument(context.Background(), "urn:li:document:doc-1")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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

func boolPtr(b bool) *bool {
	return &b
}
