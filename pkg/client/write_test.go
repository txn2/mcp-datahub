package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}

		if req.Proposal.EntityType != "dataset" {
			t.Errorf("expected entity type 'dataset', got %q", req.Proposal.EntityType)
		}
		if req.Proposal.AspectName != "editableDatasetProperties" {
			t.Errorf("expected aspect 'editableDatasetProperties', got %q", req.Proposal.AspectName)
		}

		aspectMap, ok := req.Proposal.Aspect.(map[string]any)
		if !ok {
			t.Fatal("aspect should be a map")
		}
		if aspectMap["description"] != "new description" {
			t.Errorf("expected description 'new description', got %v", aspectMap["description"])
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.UpdateDescription(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"new description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDescription_InvalidURN(t *testing.T) {
	c := &Client{logger: NopLogger{}}
	err := c.UpdateDescription(context.Background(), "not-a-urn", "desc")
	if err == nil {
		t.Fatal("expected error for invalid URN")
	}
}

func TestAddTag(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch {
		case r.Method == http.MethodGet:
			// Return existing tags
			resp := aspectResponse{
				Value: json.RawMessage(`{"tags":[{"tag":"urn:li:tag:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)

		case r.Method == http.MethodPost:
			var req ingestRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			if req.Proposal.AspectName != "globalTags" {
				t.Errorf("expected aspect 'globalTags', got %q", req.Proposal.AspectName)
			}

			// Verify the tag was added to existing tags
			aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
			var tags globalTagsAspect
			_ = json.Unmarshal(aspectBytes, &tags)
			if len(tags.Tags) != 2 {
				t.Errorf("expected 2 tags, got %d", len(tags.Tags))
			}

			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:newtag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTag_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"tags":[{"tag":"urn:li:tag:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// POST should not be called for duplicates
		t.Error("POST should not be called for duplicate tag")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTag_NoExistingTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// No existing tags
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// POST with the new tag
		var req ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
		var tags globalTagsAspect
		_ = json.Unmarshal(aspectBytes, &tags)
		if len(tags.Tags) != 1 {
			t.Errorf("expected 1 tag, got %d", len(tags.Tags))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:newtag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"tags":[{"tag":"urn:li:tag:keep"},{"tag":"urn:li:tag:remove"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		var req ingestRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
		var tags globalTagsAspect
		_ = json.Unmarshal(aspectBytes, &tags)
		if len(tags.Tags) != 1 {
			t.Errorf("expected 1 tag after removal, got %d", len(tags.Tags))
		}
		if tags.Tags[0].Tag != "urn:li:tag:keep" {
			t.Errorf("expected remaining tag 'urn:li:tag:keep', got %q", tags.Tags[0].Tag)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveTag(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:tag:remove")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddGlossaryTerm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"terms":[{"urn":"urn:li:glossaryTerm:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		var req ingestRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
		var terms glossaryTermsAspect
		_ = json.Unmarshal(aspectBytes, &terms)
		if len(terms.Terms) != 2 {
			t.Errorf("expected 2 terms, got %d", len(terms.Terms))
		}
		if req.Proposal.AspectName != "glossaryTerms" {
			t.Errorf("expected aspect 'glossaryTerms', got %q", req.Proposal.AspectName)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:newterm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddGlossaryTerm_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"terms":[{"urn":"urn:li:glossaryTerm:existing"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		t.Error("POST should not be called for duplicate term")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveGlossaryTerm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			resp := aspectResponse{
				Value: json.RawMessage(`{"terms":[{"urn":"urn:li:glossaryTerm:keep"},{"urn":"urn:li:glossaryTerm:remove"}]}`),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		var req ingestRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
		var terms glossaryTermsAspect
		_ = json.Unmarshal(aspectBytes, &terms)
		if len(terms.Terms) != 1 {
			t.Errorf("expected 1 term after removal, got %d", len(terms.Terms))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveGlossaryTerm(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"urn:li:glossaryTerm:remove")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			linkJSON := `{"elements":[{"url":"https://existing.com",` +
				`"description":"existing",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}}]}`
			resp := aspectResponse{
				Value: json.RawMessage(linkJSON),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		var req ingestRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Proposal.AspectName != "institutionalMemory" {
			t.Errorf("expected aspect 'institutionalMemory', got %q", req.Proposal.AspectName)
		}
		aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
		var memory institutionalMemoryAspect
		_ = json.Unmarshal(aspectBytes, &memory)
		if len(memory.Elements) != 2 {
			t.Errorf("expected 2 links, got %d", len(memory.Elements))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://newlink.com", "New Link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddLink_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			linkJSON := `{"elements":[{"url":"https://existing.com",` +
				`"description":"existing",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}}]}`
			resp := aspectResponse{
				Value: json.RawMessage(linkJSON),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		t.Error("POST should not be called for duplicate link")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.AddLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://existing.com", "Existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			linkJSON := `{"elements":[` +
				`{"url":"https://keep.com","description":"keep",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}},` +
				`{"url":"https://remove.com","description":"remove",` +
				`"created":{"time":0,"actor":"urn:li:corpuser:datahub"}}]}`
			resp := aspectResponse{
				Value: json.RawMessage(linkJSON),
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		var req ingestRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		aspectBytes, _ := json.Marshal(req.Proposal.Aspect)
		var memory institutionalMemoryAspect
		_ = json.Unmarshal(aspectBytes, &memory)
		if len(memory.Elements) != 1 {
			t.Errorf("expected 1 link after removal, got %d", len(memory.Elements))
		}
		if memory.Elements[0].URL != "https://keep.com" {
			t.Errorf("expected remaining link 'https://keep.com', got %q", memory.Elements[0].URL)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{
		endpoint:   server.URL + "/api/graphql",
		token:      "test-token",
		httpClient: server.Client(),
		logger:     NopLogger{},
	}

	err := c.RemoveLink(context.Background(),
		"urn:li:dataset:(urn:li:dataPlatform:hive,testdb.table,PROD)",
		"https://remove.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEntityTypeFromURN(t *testing.T) {
	tests := []struct {
		name    string
		urn     string
		want    string
		wantErr bool
	}{
		{"dataset", "urn:li:dataset:(urn:li:dataPlatform:hive,db.table,PROD)", "dataset", false},
		{"glossaryTerm", "urn:li:glossaryTerm:Classification", "glossaryTerm", false},
		{"tag", "urn:li:tag:PII", "tag", false},
		{"dashboard", "urn:li:dashboard:(looker,dashboards.123)", "dashboard", false},
		{"invalid", "not-a-urn", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := entityTypeFromURN(tt.urn)
			if (err != nil) != tt.wantErr {
				t.Errorf("entityTypeFromURN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("entityTypeFromURN() = %q, want %q", got, tt.want)
			}
		})
	}
}
