package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateDocumentContents(t *testing.T) {
	tests := []struct {
		name      string
		urn       string
		title     string
		text      string
		response  string
		wantErr   bool
		checkVars func(t *testing.T, vars map[string]any)
	}{
		{
			name:     "success with title and text",
			urn:      "urn:li:document:doc1",
			title:    "Updated Title",
			text:     "Updated body text",
			response: `{"data": {"updateDocumentContents": true}}`,
			checkVars: func(t *testing.T, vars map[string]any) {
				t.Helper()
				if vars["urn"] != "urn:li:document:doc1" {
					t.Errorf("urn = %v, want urn:li:document:doc1", vars["urn"])
				}
				input, _ := vars["input"].(map[string]any)
				if input["title"] != "Updated Title" {
					t.Errorf("title = %v, want Updated Title", input["title"])
				}
				if input["text"] != "Updated body text" {
					t.Errorf("text = %v, want Updated body text", input["text"])
				}
			},
		},
		{
			name:     "success with only title",
			urn:      "urn:li:document:doc2",
			title:    "Title Only",
			text:     "",
			response: `{"data": {"updateDocumentContents": true}}`,
			checkVars: func(t *testing.T, vars map[string]any) {
				t.Helper()
				input, _ := vars["input"].(map[string]any)
				if input["title"] != "Title Only" {
					t.Errorf("title = %v, want Title Only", input["title"])
				}
				if _, hasText := input["text"]; hasText {
					t.Error("expected no text key when text is empty")
				}
			},
		},
		{
			name:     "graphql error",
			urn:      "urn:li:document:doc1",
			title:    "Title",
			text:     "Body",
			response: `{"errors": [{"message": "document not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if tt.checkVars != nil {
					tt.checkVars(t, vars)
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
			}

			err := c.UpdateDocumentContents(context.Background(), tt.urn, tt.title, tt.text)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateDocumentStatus(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		status   string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:document:doc1",
			status:   "PUBLISHED",
			response: `{"data": {"updateDocumentStatus": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:document:doc1",
			status:   "PUBLISHED",
			response: `{"errors": [{"message": "not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
				input, _ := vars["input"].(map[string]any)
				if input["status"] != tt.status {
					t.Errorf("status = %v, want %v", input["status"], tt.status)
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
			}

			err := c.UpdateDocumentStatus(context.Background(), tt.urn, tt.status)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateDocumentRelatedEntities(t *testing.T) {
	tests := []struct {
		name       string
		urn        string
		entityURNs []string
		response   string
		wantErr    bool
	}{
		{
			name:       "success",
			urn:        "urn:li:document:doc1",
			entityURNs: []string{"urn:li:dataset:ds1", "urn:li:dataset:ds2"},
			response:   `{"data": {"updateDocumentRelatedEntities": true}}`,
		},
		{
			name:       "graphql error",
			urn:        "urn:li:document:doc1",
			entityURNs: []string{"urn:li:dataset:ds1"},
			response:   `{"errors": [{"message": "document not found"}]}`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
				input, _ := vars["input"].(map[string]any)
				urns, _ := input["entityUrns"].([]any)
				if len(urns) != len(tt.entityURNs) {
					t.Errorf("entityUrns length = %d, want %d", len(urns), len(tt.entityURNs))
				}
				for i, u := range urns {
					if u != tt.entityURNs[i] {
						t.Errorf("entityUrns[%d] = %v, want %v", i, u, tt.entityURNs[i])
					}
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
			}

			err := c.UpdateDocumentRelatedEntities(context.Background(), tt.urn, tt.entityURNs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateDocumentSubType(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		subType  string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:document:doc1",
			subType:  "KNOWLEDGE_BASE",
			response: `{"data": {"updateDocumentSubType": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:document:doc1",
			subType:  "KNOWLEDGE_BASE",
			response: `{"errors": [{"message": "not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
				}
				input, _ := vars["input"].(map[string]any)
				if input["subType"] != tt.subType {
					t.Errorf("subType = %v, want %v", input["subType"], tt.subType)
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
			}

			err := c.UpdateDocumentSubType(context.Background(), tt.urn, tt.subType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteDocument(t *testing.T) {
	tests := []struct {
		name     string
		urn      string
		response string
		wantErr  bool
	}{
		{
			name:     "success",
			urn:      "urn:li:document:doc1",
			response: `{"data": {"deleteDocument": true}}`,
		},
		{
			name:     "graphql error",
			urn:      "urn:li:document:missing",
			response: `{"errors": [{"message": "document not found"}]}`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := extractGraphQLVariables(t, r)
				if vars["urn"] != tt.urn {
					t.Errorf("urn = %v, want %v", vars["urn"], tt.urn)
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
			}

			err := c.DeleteDocument(context.Background(), tt.urn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
